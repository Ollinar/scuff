package app

import (
	"crypto"
	"log/slog"

	"github.com/Ollinar/scuff/internal/repository"
	"github.com/Ollinar/scuff/internal/search"
)

const (
	KiB = 1 << 10
	MiB = 1 << 20
	GiB = 1 << 30
)

type (
	AppOpts func(*App)
)

func WithPartialHashLength(l int64) AppOpts {
	return func(a *App) {
		if l < 1 {
			l = a.partialHashLength
		}
		a.partialHashLength = l
	}
}

func WithHasher(h crypto.Hash) AppOpts {
	return func(a *App) { a.hasher = h }
}

func WithContentDirectory(path string) AppOpts {
	return func(a *App) { a.contentDir = path }
}

func WithPluginDirectory(path string) AppOpts {
	return func(a *App) { a.pluginDir = path }
}

func WithLogger(s *slog.Logger) AppOpts {
	return func(a *App) { a.logger = s }
}

func WithArchiveRepository(repo repository.Archive) AppOpts {
	return func(a *App) { a.archiveRepo = repo }
}

func WithPageRepository(repo repository.Page) AppOpts {
	return func(a *App) { a.pageRepo = repo }
}

func WithChapterRepository(repo repository.Chapter) AppOpts {
	return func(a *App) { a.chapterRepo = repo }
}

func WithSeriesRepository(repo repository.Series) AppOpts {
	return func(a *App) { a.seriesRepo = repo }
}

func WithPluginRepository(repo repository.Plugin) AppOpts {
	return func(a *App) { a.pluginRepo = repo }
}

func WithChapterSearcher(searcher search.ChapterSearcher) AppOpts {
	return func(a *App) { a.chapterSearcher = searcher }
}
