package api

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Ollinar/scuff/internal/app"
	"github.com/Ollinar/scuff/internal/plugin"

	"github.com/labstack/echo/v4"
)

func (apiV1 APIV1) PostPlugin(c echo.Context) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	f, err := fh.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer f.Close()

	scriptB, err := io.ReadAll(f)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = apiV1.app.Plugin().Add(string(scriptB))
	if err != nil {
		if errors.Is(err, app.ErrInvalidEntity) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "file is not a valid plugin", err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (apiV1 APIV1) GetPlugins(c echo.Context) error {
	plugInf, err := apiV1.app.Plugin().GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(200, pluginInfosFromModel(plugInf))
}

func (apiV1 APIV1) RunPlugin(c echo.Context) error {
	req := ExecPluginJSON{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	plugs, err := apiV1.app.Plugin().GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	var pi *plugin.PluginInfo
	for _, v := range plugs {
		if v.Name != req.Name {
			continue
		}
		if req.Version == v.Version {
			pi = &v
			break
		} else if req.Version == "" && v.Version > pi.Version {
			pi = &v
			continue
		}

	}
	if pi == nil {
		return echo.NewHTTPError(http.StatusNotFound, "plugin not found")
	}

	err = apiV1.app.Plugin().Run(c.Request().Context(), *pi, req.Param, req.Target)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (apiv1 APIV1) UpdatePluginConfig(c echo.Context) error {
	req := PluginSetConfigJSON{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = apiv1.app.Plugin().SetConfig(c.Request().Context(),
		req.Name,
		req.Version,
		req.AutoRun,
		time.Duration(req.Delay)*time.Millisecond,
		req.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}
