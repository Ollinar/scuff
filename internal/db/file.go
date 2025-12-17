package db

import (
	"context"
	"fmt"

	"github.com/Ollinar/scuff/internal/model"

	"github.com/jmoiron/sqlx"
)

type File struct {
	acid
	db *sqlx.DB
}

func newFile(db *sqlx.DB) File {
	return File{db: db, acid: acid{db: db}}
}

// AddFiles implements repository.File.
func (f File) AddFiles(ctx context.Context, files []model.File) ([]model.File, error) {
	tx, err := f.beginTxx(ctx, f.db)
	if err != nil {
		return nil, err
	}
	defer f.rollbackTxx(ctx, tx)
	fls, err := f.addFiles(ctx, tx, files)
	if err != nil {
		return nil, err
	}
	if len(fls) == len(files) {
		return nil, fmt.Errorf("expected %d files to be added, got %d", len(files), len(fls))
	}
	err = f.commitTxx(ctx, tx)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetAllFiles implements repository.File.
func (f File) GetAllFiles(ctx context.Context) ([]model.File, error) {
	return f.getAllFiles(ctx, f.getDbtx(ctx, f.db))
}

// GetFilesByArchiveIDs implements repository.File.
func (f File) GetFilesByArchiveIDs(ctx context.Context, ids []model.ID) ([]model.File, error) {
	return f.getFilesByArchiveIds(ctx, f.getDbtx(ctx, f.db), ids)
}

// GetFilesByIDs implements repository.File.
func (f File) GetFilesByIDs(ctx context.Context, ids []model.ID) ([]model.File, error) {
	return f.getFilesByIds(ctx, f.getDbtx(ctx, f.db), ids)
}

// RemoveFiles implements repository.File.
func (f File) RemoveFiles(ctx context.Context, ids []model.ID) error {
	tx, err := f.beginTxx(ctx, f.db)
	if err != nil {
		return err
	}
	defer f.rollbackTxx(ctx, tx)

	err = f.removeFiles(ctx, tx, ids)
	if err != nil {
		return err
	}

	err = f.commitTxx(ctx, tx)
	if err != nil {
		return err
	}
	return nil
}

// Update implements repository.File.
func (f File) Update(ctx context.Context, id model.ID, file model.File) (model.File, error) {
	tx, err := f.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.File{}, err
	}
	defer rollbackTxx(tx)
	err = f.updateFile(ctx, tx, id, file)
	if err != nil {
		return model.File{}, err
	}

	err = tx.Commit()
	if err != nil {
		return model.File{}, err
	}
	return file, nil
}

func (f File) addFiles(ctx context.Context, tx *sqlx.Tx, files []model.File) ([]model.File, error) {
	if len(files) == 0 {
		return []model.File{}, nil
	}
	stm, err := tx.Preparex(
		"INSERT INTO t_file (c_path,c_modtime,c_mime,c_size,c_archiveId) VALUES (?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer stm.Close()
	for i, f := range files {
		res, err := stm.ExecContext(ctx,
			f.Path,
			f.ModTime,
			f.Mime,
			f.Size,
			f.ArchiveID,
		)
		if err != nil {
			return nil, err
		}
		files[i].ID, err = res.LastInsertId()
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func (f File) removeFiles(ctx context.Context, tx *sqlx.Tx, fileIds []model.ID) error {
	if len(fileIds) == 0 {
		return nil
	}
	q, args, err := sqlx.In("DELETE FROM t_file f WHERE f.c_id IN (?)", fileIds)
	if err != nil {
		return err
	}
	q = tx.Rebind(q)
	_, err = tx.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return nil
}

func (f File) updateFile(ctx context.Context, tx *sqlx.Tx, fID model.ID, file model.File) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE t_file f SET f.c_path = ?, f.c_modtime = ?, f.c_mime = ?, f.c_size = ? WHERE f.c_id = ?",
		file.Path, file.ModTime, file.Mime, file.Size, fID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (f File) getAllFiles(ctx context.Context, db dbtx) ([]model.File, error) {
	fls := []fileRow{}
	err := db.SelectContext(ctx, fls,
		"SELECT f.c_id,f.c_path,f.c_modtime,f.c_mime,f.c_size,f.c_archiveId, a.c_path AS c_archivePath FROM t_file f INNER JOIN t_archive a ON f.c_archiveId=a.c_id")
	if err != nil {
		return nil, err
	}
	return filerows(fls).toModel(), nil
}

func (f File) getFilesByIds(ctx context.Context, db dbtx, fIds []model.ID) ([]model.File, error) {
	if len(fIds) == 0 {
		return []model.File{}, nil
	}
	q, args, err := sqlx.In(
		"SELECT f.c_id,f.c_path,f.c_modtime,f.c_mime,f.c_size,f.c_archiveId,a.c_path AS c_archivePath FROM t_file f INNER JOIN t_archive a ON f.c_archiveId=a.c_id WHERE f.c_id IN(?)",
		fIds)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)
	fls := []fileRow{}
	err = db.SelectContext(ctx, &fls, q, args...)
	if err != nil {
		return nil, err
	}
	return filerows(fls).toModel(), nil
}

func (f File) getFilesByArchiveIds(ctx context.Context, db dbtx, arcIds []model.ID) ([]model.File, error) {
	if len(arcIds) == 0 {
		return []model.File{}, nil
	}
	q, args, err := sqlx.In(
		"SELECT f.c_id,f.c_path,f.c_modtime,f.c_mime,f.c_size,f.c_archiveId,a.c_path AS c_archivePath FROM t_file f INNER JOIN t_archive a ON f.c_archiveId=a.c_id WHERE f.c_archiveId IN (?)",
		arcIds)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)

	fls := []fileRow{}
	err = db.SelectContext(ctx, &fls, q, args...)
	if err != nil {
		return nil, err
	}
	return filerows(fls).toModel(), nil
}
