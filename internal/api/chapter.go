package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/search"

	"github.com/labstack/echo/v4"
)

func (apiV1 APIV1) GetChapters(c echo.Context) error {
	idsStr, ok := c.QueryParams()["ids"]
	ids := make([]int64, 0, len(idsStr))
	for _, v := range idsStr {
		id, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		ids = append(ids, id)
	}
	if ok {
		chaps, err := apiV1.app.Chapter().GetByIDs(c.Request().Context(), ids)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		return c.JSON(200, chaptersFromModel(chaps))
	}

	page, pageSize := 1, 10

	if p, err := strconv.Atoi(c.QueryParam("p")); err == nil {
		page = p
	}
	if ps, err := strconv.Atoi(c.QueryParam("ps")); err == nil {
		pageSize = ps
	}

	var sorting []search.Sort

	sortParam := c.QueryParam("sort")
	sortDirection := c.QueryParam("sortDirection")
	sortNamespace := c.QueryParam("sortNamespace")

	var direction search.SortDirection = search.Ascending
	if sortDirection == "desc" {
		direction = search.Descending
	}
	switch sortParam {
	case "name":
		sorting = append(sorting, search.SortByName(direction))
	case "id":
		sorting = append(sorting, search.SortByID(direction))
	case "tag":
		if sortNamespace != "" {
			prefix := strings.HasPrefix(sortNamespace, "^")
			suffx := strings.HasSuffix(sortNamespace, "$")
			nmsp := search.StringFilter{}
			if prefix && suffx {
				nmsp.Type = search.MatchingExact
				nmsp.Value = strings.TrimPrefix(strings.TrimSuffix(sortNamespace, "$"), "^")
			} else if prefix {
				nmsp.Type = search.MatchingPrefix
				nmsp.Value = strings.TrimPrefix(sortNamespace, "^")
			} else if suffx {
				nmsp.Type = search.MatchingSuffix
				nmsp.Value = strings.TrimSuffix(sortNamespace, "$")
			} else {
				nmsp.Type = search.MatchingInfix
				nmsp.Value = sortNamespace
			}

			sorting = append(sorting, search.SortByTagNamespace(direction, nmsp))
		}
	}

	if len(sorting) == 0 {
		sorting = append(sorting, search.SortByID(direction))
	}

	chaps, total, err := apiV1.app.Chapter().Search(c.Request().Context(), search.Pagination{
		Page:     &page,
		PageSize: &pageSize,
		Sorting:  sorting,
	}, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	chapsJ := chaptersFromModel(chaps)
	return c.JSON(http.StatusOK, map[string]any{
		"data":  chapsJ,
		"total": total,
	})
}

func (apiV1 APIV1) SearchChapter(c echo.Context) error {
	req := ChapterFilterJSON{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	fltr, pgnation := req.toFilterAndPagination()

	chps, total, err := apiV1.app.Chapter().Search(c.Request().Context(), *pgnation, fltr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	chapsJ := chaptersFromModel(chps)
	return c.JSON(http.StatusOK, map[string]any{
		"data":  chapsJ,
		"total": total,
	})
}

func (apiV1 APIV1) GetChapterByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	chaps, err := apiV1.app.Chapter().GetByIDs(c.Request().Context(), []model.ID{model.ID(id)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if len(chaps) < 1 {
		return echo.ErrNotFound
	}
	chap := Chapter{}
	chap.fromModel(chaps[0])
	return c.JSON(http.StatusOK, chap)
}

func (apiV1 APIV1) PostChapter(c echo.Context) error {
	reqJ := ChapterRequestJSON{}

	err := c.Bind(&reqJ)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if reqJ.Name != nil && *reqJ.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "chapter name can't be empty")
	}

	chap := model.Chapter{
		Name:    *reqJ.Name,
		PageIDs: reqJ.PageIDs,
	}

	if reqJ.Description != nil {
		chap.Descripion = *reqJ.Description
	}
	if reqJ.CoverPageID != nil {
		chap.CoverPageID = *reqJ.CoverPageID
	}

	chapPgsChk := make(map[model.ID]struct{}, len(chap.PageIDs)+1)
	for _, v := range chap.PageIDs {
		chapPgsChk[v] = struct{}{}
	}
	if chap.CoverPageID != 0 {
		chapPgsChk[chap.CoverPageID] = struct{}{}
	}

	pgsRes, err := apiV1.app.Page().GetByIDs(c.Request().Context(), append(chap.PageIDs, chap.CoverPageID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	for _, v := range pgsRes {
		delete(chapPgsChk, v.ID)
	}

	if len(chapPgsChk) != 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%d pageIds are invalid", len(chapPgsChk)))
	}

	chap.Tags, err = reqJ.parseTags()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	chapM, err := apiV1.app.Chapter().Add(c.Request().Context(), chap)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	chapJ := Chapter{}
	chapJ.fromModel(chapM)

	return c.JSON(http.StatusOK, chapJ)
}

func (apiV1 APIV1) PatchChapter(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	reqJ := ChapterRequestJSON{}

	err = c.Bind(&reqJ)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if reqJ.Name != nil && *reqJ.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "chapter name can't be empty")
	}

	tmpChps, err := apiV1.app.Chapter().GetByIDs(c.Request().Context(), []model.ID{model.ID(id)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	updates := tmpChps[0]
	if reqJ.Name != nil {
		updates.Name = *reqJ.Name
	}

	if reqJ.Description != nil {
		updates.Descripion = *reqJ.Description
	}
	if reqJ.CoverPageID != nil {
		updates.CoverPageID = *reqJ.CoverPageID
	}
	if reqJ.Tags != nil {
		updates.Tags, err = reqJ.parseTags()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
	}
	if reqJ.PageIDs != nil {
		updates.PageIDs = reqJ.PageIDs
	}

	chapM, err := apiV1.app.Chapter().Update(c.Request().Context(), model.ID(id), updates)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	chapJ := Chapter{}
	chapJ.fromModel(chapM)
	return c.JSON(http.StatusOK, chapJ)
}

func (apiV1 APIV1) DeleteChapterByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = apiV1.app.Chapter().Remove(c.Request().Context(), model.ID(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}
