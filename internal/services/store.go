package services

import (
	"context"

	"github.com/Ollinar/scuff/internal/model"
)

type ArchiveStore interface {
	Add(ctx context.Context, arc model.Archive) (model.Archive, error)
	AddFiles(ctx context.Context, files ...model.File) ([]model.File, error)
	RemoveFiles(ctx context.Context, fileIDs ...model.ID) error
	GetAll(ctx context.Context) ([]model.Archive, error)
	GetByIDs(ctx context.Context, arcIDs ...model.ID) ([]model.Archive, error)
	Remove(ctx context.Context, arcID model.ID) error
}

type FileStore interface {
	AddFiles(ctx context.Context, files ...model.File) ([]model.File, error)
	RemoveFiles(ctx context.Context, fileIDs ...model.ID) error
	GetByArchiveID(ctx context.Context, arcID model.ID) ([]model.File, error)
}
