package services

import (
	"context"
	"errors"

	"github.com/Ollinar/scuff/internal/model"
)

type ArchiveService struct {
	archiveStore ArchiveStore
	fileStore    FileStore
}

func NewArchiveService(arcSt ArchiveStore) *ArchiveService {
	return &ArchiveService{archiveStore: arcSt}
}

func (asvc ArchiveService) AddArchive(ctx context.Context, arc model.Archive) (model.Archive, error) {
	err := asvc.validateArchive(arc)
	if err != nil {
		return model.Archive{}, errors.Join(model.ErrInvalidEntity, err)
	}

	arc, err = asvc.archiveStore.Add(ctx, arc)
	if err != nil {
		return model.Archive{}, errors.Join(model.ErrUnexpected, err)
	}

	return arc, nil
}

func (asvc ArchiveService) GetAll(ctx context.Context) ([]model.Archive, error) {
	arcs, err := asvc.archiveStore.GetAll(ctx)
	if err != nil {
		return nil, errors.Join(model.ErrUnexpected, err)
	}
	return arcs, nil
}

func (asvc ArchiveService) GetByIDs(ctx context.Context, arcIDs ...model.ID) ([]model.Archive, error) {
	arcs, err := asvc.archiveStore.GetByIDs(ctx, arcIDs...)
	if err != nil {
		return nil, errors.Join(model.ErrUnexpected, err)
	}

	return arcs, nil

}

func (asvc ArchiveService) Remove(ctx context.Context, arcId model.ID) error {

	err := asvc.archiveStore.Remove(ctx, arcId)
	if err != nil {
		return errors.Join(model.ErrUnexpected, err)
	}

	return nil
}

func (asvc ArchiveService) AddFiles(ctx context.Context, files ...model.File) ([]model.File, error) {
	for _, fl := range files {
		err := asvc.validateFile(fl)
		if err != nil {
			return nil, errors.Join(model.ErrInvalidEntity, err)
		}
	}

	fls, err := asvc.archiveStore.AddFiles(ctx, files...)
	if err != nil {
		return nil, errors.Join(model.ErrUnexpected, err)
	}

	return fls, nil
}

func (asvc ArchiveService) GetArchiveFiles(ctx context.Context, arcID model.ID) ([]model.File, error) {
	fls, err := asvc.fileStore.GetByArchiveID(ctx, arcID)
	if err != nil {
		return nil, errors.Join(model.ErrUnexpected, err)
	}
	return fls, nil
}

func (asvc ArchiveService) RemoveFiles(ctx context.Context, fileIds ...model.ID) error {
	if len(fileIds) == 0 {
		return nil
	}

	err := asvc.archiveStore.RemoveFiles(ctx, fileIds...)
	if err != nil {
		return errors.Join(model.ErrUnexpected, err)
	}

	return nil
}

func (asvc ArchiveService) validateArchive(arc model.Archive) error {
	if arc.Path == "" {
		return errors.New("archive path is empty")
	}
	if arc.PartialHash == "" {
		return errors.New("archive partial hash is empty")
	}
	if arc.Type == "" {
		return errors.New("archive type is empty")
	}
	return nil
}

func (asvc ArchiveService) validateFile(fl model.File) error {
	if fl.Path == "" {
		return errors.New("file path is empty")
	}

	if fl.Mime == "" {
		return errors.New("file mime is empty")
	}

	return nil
}
