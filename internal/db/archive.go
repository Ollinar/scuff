package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ollinar/scuff/internal/model"

	"github.com/jmoiron/sqlx"
)

type Archive struct {
	acid
	File
	db *sqlx.DB
}

func newArchive(db *sqlx.DB) Archive {
	return Archive{
		db:   db,
		acid: acid{db: db},
		File: newFile(db),
	}
}

// Add implements repository.Archive.
func (a Archive) Add(ctx context.Context, arc model.Archive, files []model.File) (model.Archive, error) {
	tx, err := a.beginTxx(ctx, a.db)
	if err != nil {
		return model.Archive{}, err
	}
	defer a.rollbackTxx(ctx, tx)

	arc, err = a.addArchive(ctx, tx, arc, files)
	if err != nil {
		return model.Archive{}, err
	}

	err = a.commitTxx(ctx, tx)
	if err != nil {
		return model.Archive{}, err
	}

	return arc, nil
}

// GetByIDs implements repository.Archive.
func (a Archive) GetByIDs(ctx context.Context, ids []model.ID) ([]model.Archive, error) {
	return a.getArchivesByIds(ctx, a.getDbtx(ctx, a.db), ids)
}

// GetAll implements repository.Archive.
func (a Archive) GetAll(ctx context.Context) ([]model.Archive, error) {
	return a.getAllArchives(ctx, a.getDbtx(ctx, a.db))
}

// GetByPartialHash implements repository.Archive.
func (a Archive) GetByPartialHash(ctx context.Context, h string) ([]model.Archive, error) {
	return a.getArchivesByPartialHash(ctx, a.getDbtx(ctx, a.db), h)
}

// GetByPath implements repository.Archive.
func (a Archive) GetByPath(ctx context.Context, path string) (*model.Archive, error) {
	return a.getArchivesByPath(ctx, a.getDbtx(ctx, a.db), path)
}

// Remove implements repository.Archive.
func (a Archive) Remove(ctx context.Context, ids []model.ID) error {
	tx, err := a.beginTxx(ctx, a.db)
	if err != nil {
		return err
	}
	defer a.rollbackTxx(ctx, tx)

	err = a.removeArchives(ctx, tx, ids)
	if err != nil {
		return err
	}

	err = a.commitTxx(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

// Update implements repository.Archive.
func (a Archive) Update(ctx context.Context, id model.ID, arc model.Archive) (model.Archive, error) {
	tx, err := a.beginTxx(ctx, a.db)
	if err != nil {
		return model.Archive{}, err
	}
	defer a.rollbackTxx(ctx, tx)

	err = a.updateArchive(ctx, tx, id, arc)
	if err != nil {
		return model.Archive{}, err
	}

	err = a.commitTxx(ctx, tx)
	if err != nil {
		return model.Archive{}, err
	}

	return arc, nil
}

func (a Archive) addArchive(ctx context.Context, tx *sqlx.Tx, arc model.Archive, files []model.File) (model.Archive, error) {
	res, err := tx.ExecContext(ctx,
		"INSERT INTO  t_archive (c_path,c_size,c_modtime,c_type,c_partialhash) VALUES (?,?,?,?,?)",
		arc.Path, arc.Size, arc.ModTime, arc.Type, arc.PartialHash,
	)
	if err != nil {
		return model.Archive{}, err
	}

	arc.ID, err = res.LastInsertId()
	if err != nil {
		return model.Archive{}, err
	}

	arc.FileIds = nil
	if len(files) == 0 {
		return arc, nil
	}

	for i := range files {
		files[i].ArchiveID = arc.ID
	}
	fls, err := a.addFiles(ctx, tx, files)
	if err != nil {
		return model.Archive{}, err
	}

	if len(files) != len(fls) {
		return model.Archive{}, fmt.Errorf("expected %d files to be added, instead had %d", len(files), len(fls))
	}

	for _, fl := range fls {
		arc.FileIds = append(arc.FileIds, fl.ID)
	}

	return arc, nil
}

func (a Archive) updateArchive(ctx context.Context, tx *sqlx.Tx, arcID model.ID, newArc model.Archive) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE t_archive SET c_path = ?, c_size = ?, c_modtime = ?, c_type = ?,c_partialhash=? WHERE c_id = ?",
		newArc.Path, newArc.Size, newArc.ModTime, newArc.Type, newArc.PartialHash, arcID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (a Archive) removeArchives(ctx context.Context, tx *sqlx.Tx, ids []model.ID) error {
	if len(ids) == 0 {
		return nil
	}
	q, args, err := sqlx.In("DELETE FROM t_archive WHERE c_id IN(?)", ids)
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

func (a Archive) getAllArchives(ctx context.Context, db dbtx) ([]model.Archive, error) {
	arcs := []archiveRow{}
	err := db.SelectContext(ctx, &arcs,
		`SELECT a.c_id,a.c_path,a.c_size,a.c_modtime,a.c_type,a.c_partialhash,COALESCE(GROUP_CONCAT(af.c_id),"") AS fileIds 
		FROM t_archive a LEFT JOIN t_file af ON a.c_id = af.c_archiveId 
		GROUP BY a.c_id`,
	)
	if err != nil {
		return nil, err
	}

	return archiveRows(arcs).toModel(), err
}

func (a Archive) getArchivesByIds(ctx context.Context, db dbtx, ids []model.ID) ([]model.Archive, error) {
	if len(ids) == 0 {
		return []model.Archive{}, nil
	}

	q, args, err := sqlx.In(
		`SELECT a.c_id,a.c_path,a.c_size,a.c_modtime,a.c_type,a.c_partialhash,COALESCE(GROUP_CONCAT(af.c_id),"") AS fileIds 
		FROM t_archive a LEFT JOIN t_file af ON a.c_id = af.c_archiveId 
		WHERE a.c_id IN (?) GROUP BY a.c_id`,
		ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)

	arcs := []archiveRow{}
	err = db.SelectContext(ctx, &arcs, q, args...)
	if err != nil {
		return nil, err
	}
	return archiveRows(arcs).toModel(), nil
}

func (a Archive) getArchivesByPartialHash(ctx context.Context, db dbtx, partialHash string) ([]model.Archive, error) {
	arcs := []archiveRow{}
	err := db.SelectContext(ctx, &arcs,
		`SELECT a.c_id,a.c_path,a.c_size,a.c_modtime,a.c_type,a.c_partialhash,COALESCE(GROUP_CONCAT(af.c_id),"") AS fileIds 
		FROM t_archive a LEFT JOIN t_file af ON a.c_id = af.c_archiveId 
		WHERE a.c_partialhash=? GROUP BY a.c_id`,
		partialHash)
	if err != nil {
		return nil, err
	}
	return archiveRows(arcs).toModel(), nil
}

func (a Archive) getArchivesByPath(ctx context.Context, db dbtx, path string) (*model.Archive, error) {
	arcs := archiveRow{}
	err := db.GetContext(ctx, &arcs,
		`SELECT a.c_id,a.c_path,a.c_size,a.c_modtime,a.c_type,a.c_partialhash,COALESCE(GROUP_CONCAT(af.c_id),"") AS fileIds 
		FROM t_archive a LEFT JOIN t_file af ON a.c_id = af.c_archiveId 
		WHERE a.c_path=? GROUP BY a.c_id`,
		path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	arc := arcs.toModel()
	return &arc, nil
}
