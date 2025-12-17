package api

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Ollinar/scuff/internal/image"
	"github.com/Ollinar/scuff/internal/model"

	"github.com/labstack/echo/v4"
)

func (apiV1 APIV1) GetPageThumbnail(c echo.Context) error {
	regen := strings.ToLower(c.QueryParam("regen")) == "true"

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	pgs, err := apiV1.app.Page().GetByIDs(c.Request().Context(), []model.ID{model.ID(id)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if len(pgs) > 1 {
		return echo.ErrInternalServerError
	} else if len(pgs) == 0 {
		return echo.ErrNotFound
	}
	page := pgs[0]

	width, height := 500, 1000
	if w, err := strconv.Atoi(c.QueryParam("width")); err == nil {
		width = min(w, page.Width)
	}
	if h, err := strconv.Atoi(c.QueryParam("height")); err == nil {
		height = min(h, page.Height)
	}

	var imgFmt image.ImageType
	padding, thumbPath := generatePadedSubDir(page.ID)
	ext := ""
	switch c.QueryParam("format") {
	case "jpeg":
		imgFmt = image.JPG
		ext = ".jpg"
	case "png":
		imgFmt = image.PNG
		ext = ".png"
	case "gif":
		imgFmt = image.GIF
		ext = ".gif"
	case "jxl":
		imgFmt = image.JXL
		ext = ".jxl"
	case "webp":
		imgFmt = image.WEBP
		ext = ".webp"
	default:
		imgFmt = image.SOURCE
		ext = filepath.Ext(page.Name)

	}

	thumbPath = filepath.Join(apiV1.thumbDir, "thumb", thumbPath,
		fmt.Sprintf("%0*d_%dx%d%s", padding, page.ID, width, height, ext))
	_, err = os.Stat(thumbPath)
	if err != nil && !os.IsNotExist(err) {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if err == nil && !regen {
		return c.File(thumbPath)
	}

	err = os.MkdirAll(filepath.Dir(thumbPath), 0755)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	thumF, err := os.Create(thumbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer thumF.Close()

	err = apiV1.app.Page().GeneratePage(c.Request().Context(), thumF, page, width, height, imgFmt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = thumF.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.File(thumbPath)
}

func (apiV1 APIV1) GetPage(c echo.Context) error {
	regen := strings.ToLower(c.QueryParam("regen")) == "true"
	raw := strings.ToLower(c.QueryParam("raw")) == "true"

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	pgs, err := apiV1.app.Page().GetByIDs(c.Request().Context(), []model.ID{model.ID(id)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if len(pgs) > 1 {
		return echo.ErrInternalServerError
	} else if len(pgs) == 0 {
		return echo.ErrNotFound
	}
	page := pgs[0]

	width, height := page.Width, page.Height
	if w, err := strconv.Atoi(c.QueryParam("width")); err == nil {
		width = w
	}
	if h, err := strconv.Atoi(c.QueryParam("height")); err == nil {
		height = h
	}

	var imgFmt image.ImageType
	ext := ""
	switch c.QueryParam("format") {
	case "jpeg":
		imgFmt = image.JPG
		ext = ".jpg"
	case "png":
		imgFmt = image.PNG
		ext = ".png"
	case "gif":
		imgFmt = image.GIF
		ext = ".gif"
	case "jxl":
		imgFmt = image.JXL
		ext = ".jxl"
	case "webp":
		imgFmt = image.WEBP
		ext = ".webp"
	default:
		imgFmt = image.SOURCE
		ext = filepath.Ext(page.Name)

	}

	padding, cacheKey := generatePadedSubDir(page.ID)
	if raw {
		cacheKey = filepath.Join("cache", cacheKey, fmt.Sprintf("%0*d%s", padding, page.ID, filepath.Ext(page.Name)))
	} else {
		cacheKey = filepath.Join("cache", cacheKey,
			fmt.Sprintf("%0*d_%dx%d%s", padding, page.ID, width, height, ext))
	}
	hit, ok := apiV1.cache.Get(cacheKey)

	if ok && !regen {
		http.ServeContent(c.Response(),
			c.Request(), cacheKey,
			hit.Modtime,
			bytes.NewReader(hit.Data),
		)
		return nil
	}

	buf := bytes.NewBuffer(nil)
	if raw {
		err = apiV1.app.Page().ReadRawPage(c.Request().Context(), buf, page)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

	} else {
		err = apiV1.app.Page().GeneratePage(c.Request().Context(), buf, page, width, height, imgFmt)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}
	modtime := time.Now()

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	apiV1.cache.SetWithTTL(cacheKey,
		cachedPage{
			Modtime: modtime,
			Data:    buf.Bytes(),
		},
		int64(buf.Len()),
		apiV1.cacheTTL,
	)

	http.ServeContent(c.Response(),
		c.Request(),
		cacheKey,
		modtime,
		bytes.NewReader(buf.Bytes()),
	)
	return nil
}
