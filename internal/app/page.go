package app

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/Ollinar/scuff/internal/image"
	"github.com/Ollinar/scuff/internal/model"
)

func (ap App) Page() pageModule {
	return pageModule{app: ap}
}

type pageModule struct {
	app App
}

func (pgm pageModule) Add(ctx context.Context, fileID model.ID, pageName string, isSpread bool) (model.Page, error) {
	ok, pgs, err := pgm.generatePages(ctx, []model.ID{fileID})
	if err != nil {
		return model.Page{}, err
	}

	if !ok {
		return model.Page{}, errors.Join(ErrInvalidEntity, errors.New("fileId is either non existent or already used for other page"))
	}

	if len(pgs) != 1 {
		return model.Page{}, errors.Join(ErrUnexpected, fmt.Errorf("expected 1 page to be generated, insted got %d", len(pgs)))
	}

	page := pgs[0]
	page.Name = pageName
	page.IsSpread = isSpread

	page, err = pgm.app.pageRepo.Add(ctx, page)
	if err != nil {
		return model.Page{}, errors.Join(ErrUnexpected, err)
	}
	return page, nil
}

// AddMany will add pages from file ids with the pagename defaulting to file name, and isSpread as false.
func (pgm pageModule) AddMany(ctx context.Context, fileIds []model.ID) ([]model.Page, error) {
	ok, pgs, err := pgm.generatePages(ctx, fileIds)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Join(ErrInvalidEntity, errors.New("fileIds contains ids that either is non existent or already used for other page"))
	}

	if len(pgs) != 1 {
		return nil, errors.Join(ErrUnexpected, fmt.Errorf("expected %d page to be generated, insted got %d", len(fileIds), len(pgs)))
	}
	resP, err := pgm.app.pageRepo.AddMany(ctx, pgs)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	if len(fileIds) != len(resP) {
		return nil, errors.Join(ErrUnexpected, fmt.Errorf("expected %d pages to be added instead had %d", len(fileIds), len(resP)))
	}

	return resP, nil
}

func (pgm pageModule) GetAll(ctx context.Context) ([]model.Page, error) {
	pgs, err := pgm.app.pageRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return pgs, nil
}

func (pgm pageModule) GetByIDs(ctx context.Context, pageIDs []model.ID) ([]model.Page, error) {
	pageIDs = setFromSlice(pageIDs)
	pgs, err := pgm.app.pageRepo.GetByIDs(ctx, pageIDs)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return pgs, nil
}

func (pgm pageModule) GetByFileIDs(ctx context.Context, fileIDs []model.ID) ([]model.Page, error) {
	fileIDs = setFromSlice(fileIDs)
	pgs, err := pgm.app.pageRepo.GetByFileIDs(ctx, fileIDs)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return pgs, nil
}

func (pgm pageModule) Update(ctx context.Context, pageID model.ID, newName string, isSpread bool) (model.Page, error) {
	if newName == "" {
		return model.Page{}, errors.Join(ErrInvalidEntity, errors.New("page name can't be empty string"))
	}
	pgs, err := pgm.app.pageRepo.GetByIDs(ctx, []model.ID{pageID})
	if err != nil {
		return model.Page{}, errors.Join(ErrUnexpected, err)
	}
	if len(pgs) == 0 {
		return model.Page{}, errors.Join(ErrNotFound, fmt.Errorf("page with id %d not found", pageID))
	}
	if len(pgs) > 1 {
		return model.Page{}, errors.Join(ErrUnexpected, fmt.Errorf("expected to get 1 page, instead got %d", len(pgs)))
	}
	pg := pgs[0]

	pg.IsSpread = isSpread
	pg.Name = newName

	pg, err = pgm.app.pageRepo.Update(ctx, pageID, pg)
	if err != nil {
		return model.Page{}, errors.Join(ErrUnexpected, err)
	}

	return pg, nil
}

func (pgm pageModule) Remove(ctx context.Context, pageIds []model.ID) error {
	err := pgm.app.pageRepo.Remove(ctx, pageIds)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	return nil
}

func (pgm pageModule) AddPagesFromArchive(ctx context.Context, arcID model.ID) ([]model.Page, error) {
	arcs, err := pgm.app.Archive().GetByIDs(ctx, []model.ID{arcID})
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	if len(arcs) != 1 {
		return nil, errors.Join(ErrInvalidEntity, fmt.Errorf("expected to get 1 archive instead got %d", len(arcs)))
	}
	arc := arcs[0]

	fls, err := pgm.app.Archive().GetFilesByIDs(ctx, arc.FileIds)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	if len(fls) != len(arc.FileIds) {
		return nil, errors.Join(ErrUnexpected, fmt.Errorf("expected to get %d files, instead got %d", len(arc.FileIds), len(fls)))
	}

	existingPgs, err := pgm.app.pageRepo.GetByFileIDs(ctx, arc.FileIds)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}

	newPgs := make([]model.Page, 0, len(fls))
	finalPages := make([]model.Page, 0, len(fls))
	zr, err := zip.OpenReader(arc.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Join(ErrUnexpected, fmt.Errorf("%s is gone", arc.Path))
		}
		return nil, errors.Join(ErrUnexpected, err)
	}
	defer zr.Close()
	for _, fl := range fls {
		existIndx := slices.IndexFunc(existingPgs, func(pg model.Page) bool {
			return fl.ID == pg.FileID
		})
		// if the file is already existing, just append it to the list
		if existIndx >= 0 {
			finalPages = append(finalPages, existingPgs[existIndx])
			continue
		}
		if !pgm.app.imagePorcessor.IsSupported(fl.Mime) {
			continue
		}

		zf, err := zr.Open(fl.Path)
		if err != nil {
			return nil, errors.Join(ErrUnexpected, err)
		}
		defer zf.Close()

		img, err := pgm.app.imagePorcessor.Load(ctx, zf)
		if err != nil {
			return nil, errors.Join(ErrUnexpected, err)
		}
		defer img.Close()

		dim, err := pgm.app.imagePorcessor.Dimension(ctx, img)
		if err != nil {
			return nil, errors.Join(ErrUnexpected, err)
		}

		// using / instead of os specific sep because zip paths are always /
		pg := model.Page{
			Name:   strings.ReplaceAll(fl.Path, "/", "_"),
			FileID: fl.ID,
			File:   fl,
			Width:  dim.Width,
			Height: dim.Height,
		}
		newPgs = append(newPgs, pg)
	}

	newPgs, err = pgm.app.pageRepo.AddMany(ctx, newPgs)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	finalPages = append(finalPages, newPgs...)
	return finalPages, nil
}

func (pgm pageModule) GeneratePage(ctx context.Context, dest io.Writer, page model.Page, width int, height int, format image.ImageType) error {
	zr, err := zip.OpenReader(page.File.ArchivePath)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer zr.Close()
	zf, err := zr.Open(page.File.Path)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer zf.Close()

	img, err := pgm.app.imagePorcessor.Load(ctx, zf)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer img.Close()

	resizedImg, err := pgm.app.imagePorcessor.Resize(ctx, img, width, height, true)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer resizedImg.Close()

	err = pgm.app.imagePorcessor.Save(ctx, resizedImg, dest, format)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	return nil
}

func (pgm pageModule) ReadRawPage(ctx context.Context, dest io.Writer, page model.Page) error {
	return pgm.app.File().ReadFile(ctx, dest, page.File)
}

func (pgm pageModule) generatePages(ctx context.Context, fileIds []model.ID) (bool, []model.Page, error) {
	fls, err := pgm.app.Archive().GetFilesByIDs(ctx, fileIds)
	if err != nil {
		return false, nil, errors.Join(ErrUnexpected, err)
	}

	// make sure the result has the same length.
	// if the fls is shorter, the fileIds might have duplicated ids in fileIds.
	// if the fls is longer, there might be an anomaly to be looked at
	if len(fls) < len(fileIds) {
		// TODO: check if that realy the case
		return false, nil, nil
	} else if len(fls) > len(fileIds) {
		return false, nil, errors.Join(ErrUnexpected, fmt.Errorf("there might be a duplicate file with the same id within the fileIds"))
	}
	pgs, err := pgm.app.pageRepo.GetByFileIDs(ctx, fileIds)
	if err != nil {
		return false, nil, errors.Join(ErrUnexpected, err)
	}

	if len(pgs) > 0 {
		return false, nil, nil
	}

	newPages := make([]model.Page, 0, len(fileIds))
	archiveIds := make([]model.ID, 0, len(fileIds))
	fileMap := make(map[model.ID]model.File, len(fileIds))
	for _, fileID := range fileIds {
		// make sure each id exist
		flIdx := slices.IndexFunc(fls, func(el model.File) bool { return el.ID == fileID })
		if flIdx < 0 {
			return false, nil, nil
		}

		fl := fls[flIdx]
		fileMap[fileID] = fl
	}

	// TODO: wtf is this?
	for _, fl := range fls {
		fileMap[fl.ID] = fl
		exist := slices.Contains(archiveIds, fl.ArchiveID)
		if !exist {
			archiveIds = append(archiveIds, fl.ArchiveID)
		}
	}

	arcs, err := pgm.app.archiveRepo.GetByIDs(ctx, archiveIds)
	if err != nil {
		return false, nil, errors.Join(ErrUnexpected, err)
	}
	if len(arcs) != len(archiveIds) {
		return false, nil, errors.Join(ErrUnexpected, fmt.Errorf("expected to get %d archive, instead got %d", len(archiveIds), len(arcs)))
	}

	for _, arc := range arcs {
		arcR, err := zip.OpenReader(arc.Path)
		if err != nil {
			return false, nil, errors.Join(ErrUnexpected, err)
		}
		defer arcR.Close()
		for _, fID := range arc.FileIds {
			fl, pending := fileMap[fID]
			if !pending {
				continue
			}

			arcFlR, err := arcR.Open(fl.Path)
			if err != nil {
				return false, nil, errors.Join(ErrUnexpected, err)
			}
			defer arcFlR.Close()

			img, err := pgm.app.imagePorcessor.Load(ctx, arcFlR)
			if err != nil {
				return false, nil, err
			}
			defer img.Close()

			ok := pgm.app.imagePorcessor.IsSupported(fl.Mime)
			if !ok {
				return false, nil, errors.Join(ErrUnexpected, fmt.Errorf("format file with id of %d is unsupported", fl.ID))
			}

			dim, err := pgm.app.imagePorcessor.Dimension(ctx, img)
			if err != nil {
				return false, nil, errors.Join(ErrUnexpected, err)
			}
			newPages = append(newPages, model.Page{
				Name:   strings.ReplaceAll(fl.Path, "/", "_"),
				Width:  dim.Width,
				Height: dim.Height,
				FileID: fl.ID,
				File:   fl,
			})

		}

	}

	return true, newPages, nil
}
