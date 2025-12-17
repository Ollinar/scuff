// Package repository defines the interface between the application and data persistence
package repository

import (
	"context"

	"github.com/Ollinar/scuff/internal/model"
)

type Acid interface {
	// WithTransaction should mark the context for acidity
	WithTransaction(ctx context.Context) (context.Context, error)
	Save(ctx context.Context) error
	// Rollback should be a no-op if the ctx is already saved
	Rollback(ctx context.Context) error
}

type Archive interface {
	File

	// Add should ignore the fileIds field
	Add(ctx context.Context, archive model.Archive, files []model.File) (model.Archive, error)
	GetAll(ctx context.Context) ([]model.Archive, error)
	GetByIDs(ctx context.Context, ids []model.ID) ([]model.Archive, error)
	GetByPath(ctx context.Context, path string) (*model.Archive, error)
	GetByPartialHash(ctx context.Context, h string) ([]model.Archive, error)
	// GetByFileIDs should return the archives with atleast the fileId field containing the matching fileId passed as paramerter
	// GetByFileIDs(ctx context.Context, ids []model.ID) ([]model.Archive, error)
	// Update should ignore the fileIds field
	Update(ctx context.Context, archiveID model.ID, updated model.Archive) (model.Archive, error)
	Remove(ctx context.Context, archiveIDs []model.ID) error
}

type File interface {
	AddFiles(ctx context.Context, files []model.File) ([]model.File, error)
	GetAllFiles(ctx context.Context) ([]model.File, error)
	GetFilesByIDs(ctx context.Context, ids []model.ID) ([]model.File, error)
	GetFilesByArchiveIDs(ctx context.Context, ids []model.ID) ([]model.File, error)
	RemoveFiles(ctx context.Context, files []model.ID) error
}

type Page interface {
	Add(ctx context.Context, page model.Page) (model.Page, error)
	AddMany(ctx context.Context, pages []model.Page) ([]model.Page, error)
	GetAll(ctx context.Context) ([]model.Page, error)
	GetByIDs(ctx context.Context, ids []model.ID) ([]model.Page, error)
	GetByFileIDs(ctx context.Context, ids []model.ID) ([]model.Page, error)
	Update(ctx context.Context, id model.ID, page model.Page) (model.Page, error)
	Remove(ctx context.Context, ids []model.ID) error
}

type Chapter interface {
	Add(ctx context.Context, chap model.Chapter) (model.Chapter, error)
	GetAll(ctx context.Context) ([]model.Chapter, error)
	GetByIDs(ctx context.Context, ids []model.ID) ([]model.Chapter, error)
	Update(ctx context.Context, id model.ID, chap model.Chapter) (model.Chapter, error)
	Remove(ctx context.Context, id model.ID) error
}

type Series interface {
	Add(ctx context.Context, series model.Series) (model.Series, error)
	GetAll(ctx context.Context) ([]model.Series, error)
	GetByIDs(ctx context.Context, ids []model.ID) ([]model.Series, error)
	Update(ctx context.Context, id model.ID, series model.Series) (model.Series, error)
	Remove(ctx context.Context, id model.ID) error
}

type Plugin interface {
	GetConfig(ctx context.Context, name string, ver string) (map[string]string, error)
	StoreConfig(ctx context.Context, name string, ver string, conf map[string]string) error
}
