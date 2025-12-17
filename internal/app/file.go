package app

import (
	"archive/zip"
	"context"
	"errors"
	"io"

	"github.com/Ollinar/scuff/internal/model"
)

func (ap App) File() fileModule {
	return fileModule{app: ap}
}

type fileModule struct {
	app App
}

func (fm fileModule) GetFilesByArchiveID(ctx context.Context, arcID model.ID) ([]model.File, error) {
	fl, err := fm.app.archiveRepo.GetFilesByArchiveIDs(ctx, []model.ID{arcID})
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return fl, nil
}

func (fm fileModule) ReadFile(ctx context.Context, dest io.Writer, file model.File) error {
	arcR, err := zip.OpenReader(file.ArchivePath)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer arcR.Close()
	arcF, err := arcR.Open(file.Path)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer arcF.Close()
	_, err = io.Copy(dest, arcF)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	return nil
}
