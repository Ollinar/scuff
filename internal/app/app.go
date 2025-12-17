// Package app contains is the main application
package app

import (
	"context"
	"crypto"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Ollinar/scuff/internal/image"
	"github.com/Ollinar/scuff/internal/plugin"
	"github.com/Ollinar/scuff/internal/repository"
	"github.com/Ollinar/scuff/internal/search"
	"github.com/Ollinar/scuff/internal/vips"

	"github.com/fsnotify/fsnotify"
)

func NewApp(ctx context.Context, dns string, plugingProvider plugin.Provider, opts ...AppOpts) (*App, error) {
	ap := &App{
		logger:            slog.Default(),
		hasher:            crypto.SHA256,
		partialHashLength: MiB,
		pluginProvider:    plugingProvider,
	}
	for _, fn := range opts {
		fn(ap)
	}
	ap.imagePorcessor = vips.NewVipsImageProcessor(ap.logger)

	err := ap.imagePorcessor.Start(ctx)
	if err != nil {
		return nil, err
	}

	err = ap.setupWatcher()
	if err != nil {
		return nil, err
	}

	return ap, nil
}

type App struct {
	// configs
	logger            *slog.Logger
	hasher            crypto.Hash
	partialHashLength int64
	contentDir        string
	pluginDir         string

	watcher        *fsnotify.Watcher
	imagePorcessor image.Processor
	pluginProvider plugin.Provider

	// repos
	archiveRepo repository.Archive
	pageRepo    repository.Page
	chapterRepo repository.Chapter
	seriesRepo  repository.Series
	pluginRepo  repository.Plugin

	chapterSearcher search.ChapterSearcher
}

func (ap App) Close() error {
	err := ap.imagePorcessor.Shutdown()
	if err != nil {
		return err
	}
	return ap.watcher.Close()
}

func (ap *App) setupWatcher() error {
	wtchr, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	ap.watcher = wtchr
	err = fs.WalkDir(os.DirFS(ap.contentDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			err = ap.watcher.Add(filepath.Join(ap.contentDir, path))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	go watcherHandler(ap)

	return nil
}
