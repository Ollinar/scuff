package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Ollinar/scuff/internal/api"
	"github.com/Ollinar/scuff/internal/app"
	"github.com/Ollinar/scuff/internal/blevesearcher"
	"github.com/Ollinar/scuff/internal/db"
	"github.com/Ollinar/scuff/internal/luaplugin"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

// TODO: implement a graceful shutdown
func main() {
	conf := loadConfig()

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := db.NewSqlite(conf.DBDNS)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	chapterSearcher, isNew, err := blevesearcher.NewBleveSeacher(conf.chapterIndexPath)
	if err != nil {
		panic(err)
	}

	if isNew {
		go func() {
			chaps, err := db.Chapter().GetAll(appCtx)
			if err != nil {
				// for now ignore it and let the index be empty
				return
			}
			chapterSearcher.IndexChapters(appCtx, chaps...)
		}()
	}

	logggerOpts := slog.HandlerOptions{
		Level: slog.LevelWarn,
	}
	appLogger := slog.New(slog.NewTextHandler(os.Stdout, &logggerOpts))

	pluginProv := luaplugin.NewLuaPlugin(appLogger.WithGroup("PLUGIN"))
	if conf.Debug {
		logggerOpts.Level = slog.LevelDebug
		logggerOpts.AddSource = true
	}

	ap, err := app.NewApp(appCtx,
		conf.DBDNS, pluginProv,
		app.WithContentDirectory(conf.ContentDir),
		app.WithLogger(appLogger),
		app.WithPluginDirectory(conf.PluginDir),
		app.WithArchiveRepository(db.Archive()),
		app.WithPageRepository(db.Page()),
		app.WithChapterRepository(db.Chapter()),
		app.WithSeriesRepository(db.Series()),
		app.WithPluginRepository(db.Plugin()),
		app.WithChapterSearcher(chapterSearcher),
	)
	if err != nil {
		panic(err)
	}
	defer ap.Close()

	pluginProv.AddModule(
		luaplugin.WithJSONModule(),
		luaplugin.WithHTTPModule(&http.Client{}),
		luaplugin.WithUtilModule(),
		luaplugin.WithAppModlue(ap),
	)

	e := echo.New()
	e.Debug = true
	e.Use(middleware.CORS(), middleware.Gzip())

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	apiv1 := api.NewAPIV1(ap, conf.ThumbDir, conf.CacheLifetime, int(conf.MaxCacheSize))
	setupRoute(e, apiv1)
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  conf.WebappDir,
		HTML5: true,
	}))

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", conf.Host, conf.Port)))
}

func setupRoute(e *echo.Echo, apiV1 api.APIV1) {
	eV1 := e.Group("/api/v1")
	eV1.GET("/archive/", apiV1.GetAllArchive)
	eV1.POST("/archive/scan", apiV1.ArchiveScan)
	eV1.GET("/archive/:id", apiV1.GetArchiveByID)
	eV1.GET("/archive/:id/file", apiV1.GetArchiveFiles)
	eV1.GET("/archive/:id/page", apiV1.GetArchivePages)
	eV1.POST("/archive/:id/page/generate", apiV1.ArchiveGeneratePage)
	eV1.POST("/archive/:id/chapter/generate", apiV1.ArchiveGenerateChapter)

	eV1.GET("/chapter/", apiV1.GetChapters)
	eV1.GET("/chapter/:id/", apiV1.GetChapterByID)
	eV1.POST("/chapter/search", apiV1.SearchChapter)
	eV1.POST("/chapter/", apiV1.PostChapter)
	eV1.PATCH("/chapter/:id/", apiV1.PatchChapter)
	eV1.DELETE("/chapter/:id/", apiV1.DeleteChapterByID)

	eV1.GET("/page/:id/", apiV1.GetPage)
	eV1.GET("/page/:id/info/", nil)
	eV1.GET("/page/:id/thumbnail/", apiV1.GetPageThumbnail)

	eV1.GET("/series", nil)
	eV1.GET("/series/:id", nil)
	eV1.POST("/series", nil)
	eV1.PATCH("/series/:id", nil)
	eV1.DELETE("/series/:id", nil)

	eV1.POST("/plugin/upload", apiV1.PostPlugin)
	eV1.GET("/plugin/", apiV1.GetPlugins)
	eV1.POST("/plugin/", apiV1.RunPlugin)
	eV1.PATCH("/plugin/", apiV1.UpdatePluginConfig)

	eV1.GET("/series/:id/cover", nil)
	eV1.GET("/file", apiV1.GetArchiveFiles)
}

func loadConfig() Config {
	viper.SetDefault("app.debug", false)
	viper.SetDefault("app.dbDSN", "./app.db?_fk=true&_journal=wal")
	viper.SetDefault("app.webappDir", "./public")
	viper.SetDefault("app.contentDir", "./content")
	viper.SetDefault("app.thumbDir", "./thumb")
	viper.SetDefault("app.pluginDir", "./plugins")
	viper.SetDefault("app.chapterIndex", "./chapterIndex.bleve")
	viper.SetDefault("app.maxCacheSize", 500)
	viper.SetDefault("app.cacheLifetime", time.Minute*10)
	viper.SetDefault("app.cacheDir", "./cache")
	viper.SetDefault("app.host", "127.0.0.1")
	viper.SetDefault("app.port", 6996)

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err.Error())
		}
	}

	viper.SetEnvPrefix("scuff")
	viper.EnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	return Config{
		Debug:            viper.GetBool("app.debug"),
		DBDNS:            viper.GetString("app.dbDSN"),
		WebappDir:        viper.GetString("app.webappDir"),
		ContentDir:       viper.GetString("app.contentDir"),
		ThumbDir:         viper.GetString("app.thumbDir"),
		PluginDir:        viper.GetString("app.pluginDir"),
		chapterIndexPath: viper.GetString("app.chapterIndex"),
		MaxCacheSize:     viper.GetUint("app.maxCacheSize"),
		CacheLifetime:    viper.GetDuration("app.cacheLifetime"),
		CacheDir:         viper.GetString("app.cacheDir"),
		Host:             viper.GetString("app.host"),
		Port:             viper.GetInt("app.port"),
	}
}

type Config struct {
	Debug            bool
	DBDNS            string
	WebappDir        string
	TmpDir           string
	ContentDir       string
	ThumbDir         string
	CacheDir         string
	MaxCacheSize     uint
	CacheLifetime    time.Duration
	PluginDir        string
	chapterIndexPath string
	Host             string
	Port             int
}
