package api

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Ollinar/scuff/internal/app"
	"github.com/Ollinar/scuff/internal/model"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/labstack/echo/v4"
)

type APIV1 struct {
	app      *app.App
	thumbDir string
	cache    *ristretto.Cache[string, cachedPage]
	cacheTTL time.Duration

	scanInProgress *atomic.Bool
}

func NewAPIV1(app *app.App, thumnailDir string, cacheTTL time.Duration, maxCacheSize int) APIV1 {
	cacheT, err := ristretto.NewCache(&ristretto.Config[string, cachedPage]{
		NumCounters: 1e7,
		BufferItems: 64,
		MaxCost:     int64(maxCacheSize * 1024 * 1024),
	})
	if err != nil {
		panic(err)
	}

	return APIV1{
		app: app, thumbDir: thumnailDir,
		scanInProgress: &atomic.Bool{},
		cache:          cacheT,
		cacheTTL:       cacheTTL,
	}
}

func (apiV1 APIV1) GetAllArchive(c echo.Context) error {
	arcsM, err := apiV1.app.Archive().GetAll(c.Request().Context())
	if err != nil {
		return err
	}
	arcs := archivesFromModel(arcsM)

	return c.JSON(http.StatusOK, arcs)
}

func (apiV1 APIV1) GetArchiveByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}

	arcs, err := apiV1.app.Archive().GetByIDs(c.Request().Context(), []model.ID{model.ID(id)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if len(arcs) < 1 {
		return echo.ErrNotFound
	}
	arc := arcs[0]
	arcJ := Archive{}
	arcJ.fromModel(arc)

	return c.JSON(http.StatusOK, arcJ)
}

func (apiV1 APIV1) GetArchiveFiles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	fls, err := apiV1.app.File().GetFilesByArchiveID(c.Request().Context(), model.ID(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	flsJ := filesFromModel(fls)

	return c.JSON(http.StatusOK, flsJ)
}

func (apiV1 APIV1) GetArchivePages(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	arcs, err := apiV1.app.Archive().GetByIDs(c.Request().Context(), []model.ID{model.ID(id)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	var pgsJ []Page

	if len(arcs) < 1 {
		return c.JSON(http.StatusOK, pgsJ)
	}

	pgs, err := apiV1.app.Page().GetByFileIDs(c.Request().Context(), arcs[0].FileIds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	pgsJ = pagesFromModel(pgs)
	return c.JSON(http.StatusOK, pgsJ)
}

func (apiV1 APIV1) ArchiveGeneratePage(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	pgs, err := apiV1.app.Page().AddPagesFromArchive(c.Request().Context(), model.ID(id))
	if err != nil {
		return err
	}
	pages := pagesFromModel(pgs)

	return c.JSON(http.StatusOK, pages)
}

func (apiV1 APIV1) ArchiveGenerateChapter(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	chp, err := apiV1.app.Chapter().AddChapterFromArchive(c.Request().Context(), model.ID(id))
	if err != nil {
		return err
	}

	chapter := Chapter{}
	chapter.fromModel(chp)

	return c.JSON(http.StatusOK, chapter)
}

func (apiV1 APIV1) ArchiveScan(c echo.Context) error {
	if apiV1.scanInProgress.Load() {
		return c.NoContent(http.StatusOK)
	}

	apiV1.scanInProgress.Store(true)
	go func() {
		err := apiV1.app.Archive().ScanContentDirectory(context.Background())
		if err != nil {
			slog.Error(err.Error())
		}
		apiV1.scanInProgress.Store(false)
	}()
	return c.NoContent(http.StatusOK)
}
