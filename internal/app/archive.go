package app

import (
	"cmp"
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/plugin"
	"github.com/Ollinar/scuff/pkg/archive"

	"github.com/gabriel-vasile/mimetype"
)

func (ap App) Archive() archiveModule {
	return archiveModule{ap}
}

type archiveModule struct {
	app App
}

func (am archiveModule) Add(ctx context.Context, arc model.Archive, files []model.File) (model.Archive, error) {
	err := am.validateArchive(arc)
	if err != nil {
		return model.Archive{}, errors.Join(ErrInvalidEntity, err)
	}

	for _, fl := range files {
		err := am.validateFile(fl)
		if err != nil {
			return model.Archive{}, errors.Join(ErrInvalidEntity, err)
		}
	}

	arc, err = am.app.archiveRepo.Add(ctx, arc, files)
	if err != nil {
		return model.Archive{}, errors.Join(ErrUnexpected, err)
	}

	return arc, nil
}

func (am archiveModule) GetAll(ctx context.Context) ([]model.Archive, error) {
	arcs, err := am.app.archiveRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return arcs, nil
}

func (am archiveModule) GetByIDs(ctx context.Context, archiveIDs []model.ID) ([]model.Archive, error) {
	arcs, err := am.app.archiveRepo.GetByIDs(ctx, archiveIDs)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return arcs, nil
}

func (am archiveModule) GetByPartialHash(ctx context.Context, partialHash string) ([]model.Archive, error) {
	arcs, err := am.app.archiveRepo.GetByPartialHash(ctx, partialHash)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return arcs, nil
}

func (am archiveModule) GetByPath(ctx context.Context, path string) (*model.Archive, error) {
	arcs, err := am.app.archiveRepo.GetByPath(ctx, path)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return arcs, nil
}

func (am archiveModule) Update(ctx context.Context, arcID model.ID, updated model.Archive) (model.Archive, error) {
	err := am.validateArchive(updated)
	if err != nil {
		return model.Archive{}, errors.Join(ErrInvalidEntity, err)
	}
	updated.ID = arcID
	arc, err := am.app.archiveRepo.Update(ctx, arcID, updated)
	if err != nil {
		return model.Archive{}, errors.Join(ErrUnexpected, err)
	}

	return arc, nil
}

func (am archiveModule) Remove(ctx context.Context, arcIds []model.ID) error {
	if len(arcIds) == 0 {
		return nil
	}
	var err error
	ctx, err = beginTransaction(ctx, am.app.archiveRepo)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer rollbackTransaction(ctx, am.app.archiveRepo)

	arcs, err := am.app.archiveRepo.GetByIDs(ctx, arcIds)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	// assuming each arcs have 10 files, make a buffer
	fIds := make([]model.ID, 0, len(arcs)*10)
	for _, arc := range arcs {
		fIds = append(fIds, arc.FileIds...)
	}

	err = am.RemoveFiles(ctx, fIds)
	if err != nil {
		return err
	}

	err = am.app.archiveRepo.Remove(ctx, arcIds)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	err = saveTransaction(ctx, am.app.archiveRepo)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	return nil
}

func (am archiveModule) AddFiles(ctx context.Context, files []model.File) ([]model.File, error) {
	if len(files) == 0 {
		return files, nil
	}

	for _, fl := range files {
		err := am.validateFile(fl)
		if err != nil {
			return nil, errors.Join(ErrInvalidEntity, err)
		}
		if fl.ArchiveID == 0 {
			return nil, errors.Join(ErrInvalidEntity, errors.New("file's archive id is 0"))
		}
	}
	var err error

	files, err = am.app.archiveRepo.AddFiles(ctx, files)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}

	return files, nil
}

func (am archiveModule) RemoveFiles(ctx context.Context, fileIds []model.ID) error {
	if len(fileIds) == 0 {
		return nil
	}

	err := am.app.archiveRepo.RemoveFiles(ctx, fileIds)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	return nil
}

func (am archiveModule) GetFilesByIDs(ctx context.Context, fileIds []model.ID) ([]model.File, error) {
	fls, err := am.app.archiveRepo.GetFilesByIDs(ctx, fileIds)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return fls, nil
}

func (am archiveModule) AddZipToArchives(ctx context.Context, zPath string) (model.Archive, bool, error) {
	arcChk, err := am.GetByPath(ctx, zPath)
	if err != nil {
		return model.Archive{}, false, err
	}
	// NOTE: for now, rely on modtime for detecting content change, even though it might not be reliable.
	if arcChk != nil {
		inf, err := os.Stat(zPath)
		if err != nil {
			return model.Archive{}, false, errors.Join(ErrUnexpected, err)
		}
		if arcChk.ModTime == inf.ModTime().UnixMilli() {
			return *arcChk, false, nil
		}

	}

	pArc, err := archive.ArchiveFromZip(zPath)
	if err != nil {
		return model.Archive{}, false, errors.Join(ErrUnexpected, err)
	}
	var arc model.Archive
	arc.Type = pArc.Type
	arc.Path = zPath
	arc.Size = pArc.Size
	arc.ModTime = pArc.ModTime.UnixMilli()

	zFiles := make([]model.File, 0, len(pArc.Files))
	for _, zf := range pArc.Files {
		zFiles = append(zFiles, model.File{
			Path:        zf.Path,
			Mime:        zf.Mime,
			ModTime:     zf.ModTime.UnixMilli(),
			Size:        zf.Size,
			ArchivePath: arc.Path,
		})
	}

	if arcChk != nil && len(arcChk.FileIds) == len(zFiles) {
		arcChkFls, err := am.GetFilesByIDs(ctx, arcChk.FileIds)
		if err != nil {
			return model.Archive{}, false, err
		}
		slices.SortFunc(zFiles, func(a, b model.File) int { return cmp.Compare(a.Path, b.Path) })
		slices.SortFunc(arcChkFls, func(a, b model.File) int { return cmp.Compare(a.Path, b.Path) })
		eq := slices.EqualFunc(arcChkFls, zFiles, func(e1, e2 model.File) bool {
			return e1.Mime == e2.Mime && e1.Path == e2.Path && e1.ModTime == e2.ModTime && e1.Size == e2.Size
		})
		if eq {
			return *arcChk, false, nil
		}
		// if the contents of the archive is not the same, assume its been replaced. delete it and make a new one
		err = am.Remove(ctx, []model.ID{arcChk.ID})
		if err != nil {
			return model.Archive{}, false, err
		}
	}

	f, err := os.Open(zPath)
	if err != nil {
		return model.Archive{}, false, errors.Join(ErrUnexpected, err)
	}
	defer f.Close()
	arc.PartialHash, err = generatePartialHash(f, am.app.hasher, am.app.partialHashLength)
	if err != nil {
		return model.Archive{}, false, errors.Join(ErrUnexpected, err)
	}

	// chack if there are achives with same partial hash.
	// if there are, check if its still there, if its not, check if the new archive has the same file contents.
	// if it does, assume that the archive was moved and update the path to point to currnet one.
	res, err := am.GetByPartialHash(ctx, arc.PartialHash)
	if err != nil {
		return model.Archive{}, false, err
	}

	if len(res) == 0 {
		arc, err := am.Add(ctx, arc, zFiles)
		if err != nil {
			return model.Archive{}, false, err
		}
		return arc, true, nil
	}

	for _, exstingArc := range res {

		existingArcFls, err := am.GetFilesByIDs(ctx, exstingArc.FileIds)
		if err != nil {
			return model.Archive{}, false, err
		}

		slices.SortFunc(zFiles, func(a, b model.File) int { return cmp.Compare(a.Path, b.Path) })
		slices.SortFunc(existingArcFls, func(a, b model.File) int { return cmp.Compare(a.Path, b.Path) })
		eq := slices.EqualFunc(existingArcFls, zFiles, func(e1, e2 model.File) bool {
			return e1.Mime == e2.Mime && e1.Path == e2.Path && e1.ModTime == e2.ModTime && e1.Size == e2.Size
		})
		// check if the existing archive is gone (possibly moved to the current one)
		_, err = os.Stat(exstingArc.Path)
		if err != nil && !os.IsNotExist(err) {
			return model.Archive{}, false, errors.Join(ErrUnexpected, err)
		}

		if os.IsNotExist(err) && eq {
			// assume its moved, update the existing one.
			arc, err = am.Update(ctx, exstingArc.ID, arc)
			if err != nil {
				return model.Archive{}, false, err
			}
			return arc, false, nil
		} else if os.IsNotExist(err) {
			continue
		}

	}

	arc, err = am.Add(ctx, arc, zFiles)
	if err != nil {
		return model.Archive{}, false, err
	}

	return arc, true, nil
}

func (am archiveModule) ScanContentDirectory(ctx context.Context) error {
	arcs, err := am.GetAll(ctx)
	if err != nil {
		return err
	}
	arcMap := make(map[string]model.Archive, len(arcs))
	for _, v := range arcs {
		arcMap[v.Path] = v
	}

	zipPaths := make([]string, 0)

	err = fs.WalkDir(os.DirFS(am.app.contentDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		fullpath := filepath.Join(am.app.contentDir, path)

		m, err := mimetype.DetectFile(fullpath)
		if err != nil {
			return err
		}

		// TODO: mkae other format of archive suppored
		if !m.Is("application/zip") {
			return nil
		}

		zipPaths = append(zipPaths, fullpath)

		return nil
	})
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	wg := sync.WaitGroup{}
	limiterCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	plugins, err := am.app.Plugin().LoadAutoRuns(ctx, plugin.TargetArchive)
	if err != nil {
		return err
	}

	for _, zp := range zipPaths {
		limiterCh <- struct{}{}
		wg.Add(1)
		go func(path string) {
			defer func() {
				<-limiterCh
				wg.Done()
				am.app.logger.Debug("finished processing archive path", slog.String("path", path))
			}()
			select {
			case <-cancelCtx.Done():
				return
			default:
			}
			am.app.logger.Debug("processing archive path", slog.String("path", path))
			arc, isNew, err := am.AddZipToArchives(cancelCtx, path)
			if err != nil {
				errCh <- err
				cancel()
			}
			if isNew {
				for _, plug := range plugins {
					plug.QueueUp(arc.ID)
				}
			}
		}(zp)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (am archiveModule) validateArchive(arc model.Archive) error {
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

func (am archiveModule) validateFile(fl model.File) error {
	if fl.Path == "" {
		return errors.New("file path is empty")
	}

	if fl.Mime == "" {
		return errors.New("file mime is empty")
	}

	return nil
}
