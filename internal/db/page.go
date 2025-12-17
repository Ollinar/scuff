package db

import (
	"context"
	"fmt"

	"github.com/Ollinar/scuff/internal/model"

	"github.com/jmoiron/sqlx"
)

type Page struct {
	acid
	db *sqlx.DB
}

func newPage(db *sqlx.DB) Page {
	return Page{db: db, acid: acid{db: db}}
}

// Add implements repository.Page.
func (p Page) Add(ctx context.Context, page model.Page) (model.Page, error) {
	tx, err := p.beginTxx(ctx, p.db)
	if err != nil {
		return model.Page{}, err
	}
	defer p.rollbackTxx(ctx, tx)

	pgs, err := p.addPages(ctx, tx, []model.Page{page})
	if err != nil {
		return model.Page{}, err
	}
	// should never happen, but just incase
	if len(pgs) != 1 {
		return model.Page{}, fmt.Errorf("expected to return 1 page, instead got %d", len(pgs))
	}
	page = pgs[0]

	err = p.commitTxx(ctx, tx)
	if err != nil {
		return model.Page{}, err
	}
	return page, nil
}

func (p Page) AddMany(ctx context.Context, pages []model.Page) ([]model.Page, error) {
	tx, err := p.beginTxx(ctx, p.db)
	if err != nil {
		return nil, err
	}
	defer p.rollbackTxx(ctx, tx)

	pgs, err := p.addPages(ctx, tx, pages)
	if err != nil {
		return nil, err
	}

	// should never happen, but just incase
	if len(pgs) != len(pages) {
		return nil, fmt.Errorf("expecter %d pages, got %d", len(pgs), len(pages))
	}
	pages = pgs

	err = p.commitTxx(ctx, tx)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

// GetAll implements repository.Page.
func (p Page) GetAll(ctx context.Context) ([]model.Page, error) {
	return p.getAllPages(ctx, p.getDbtx(ctx, p.db))
}

// GetByFileIDs implements repository.Page.
func (p Page) GetByFileIDs(ctx context.Context, ids []model.ID) ([]model.Page, error) {
	return p.getPagesByFileIds(ctx, p.getDbtx(ctx, p.db), ids)
}

// GetByIDs implements repository.Page.
func (p Page) GetByIDs(ctx context.Context, ids []model.ID) ([]model.Page, error) {
	return p.getPagesByIds(ctx, p.getDbtx(ctx, p.db), ids)
}

// Remove implements repository.Page.
func (p Page) Remove(ctx context.Context, ids []model.ID) error {
	tx, err := p.beginTxx(ctx, p.db)
	if err != nil {
		return err
	}
	defer p.rollbackTxx(ctx, tx)
	err = p.removePages(ctx, tx, ids)
	if err != nil {
		return err
	}

	err = p.commitTxx(ctx, tx)
	if err != nil {
		return err
	}
	return nil
}

// Update implements repository.Page.
func (p Page) Update(ctx context.Context, id model.ID, page model.Page) (model.Page, error) {
	tx, err := p.beginTxx(ctx, p.db)
	if err != nil {
		return model.Page{}, err
	}
	defer rollbackTxx(tx)

	err = p.updatePage(ctx, tx, id, page)
	if err != nil {
		return model.Page{}, err
	}

	err = p.commitTxx(ctx, tx)
	if err != nil {
		return model.Page{}, err
	}

	return page, nil
}

func (p Page) addPages(ctx context.Context, tx *sqlx.Tx, pages []model.Page) ([]model.Page, error) {
	if len(pages) == 0 {
		return []model.Page{}, nil
	}
	stm, err := tx.PreparexContext(ctx,
		"INSERT INTO t_page (c_width,c_height,c_isSpread,c_fileId,c_pageName) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer stm.Close()
	for i, page := range pages {
		res, err := stm.ExecContext(ctx,
			page.Width, page.Height, page.IsSpread, page.FileID, page.Name)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		pages[i].ID = id

	}
	return pages, nil
}

func (p Page) removePages(ctx context.Context, tx *sqlx.Tx, pageIds []model.ID) error {
	if len(pageIds) == 0 {
		return nil
	}
	q, args, err := sqlx.In("DELETE FROM t_pages WHERE c_id IN (?)", pageIds)
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

func (p Page) updatePage(ctx context.Context, tx *sqlx.Tx, pageID model.ID, page model.Page) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE t_page SET c_width=?,c_height=?,c_isSpread=?,c_fileId=?,c_pageName=? WHERE c_id=?",
		page.Width, page.Height, page.IsSpread, page.FileID, page.Name, pageID)
	if err != nil {
		return err
	}

	return nil
}

func (p Page) getAllPages(ctx context.Context, db dbtx) ([]model.Page, error) {
	var pgs []pageRow
	err := db.SelectContext(ctx, &pgs,
		`SELECT p.c_id, p.c_width, p.c_height, p.c_isSpread, 
		p.c_fileId, p.c_pageName,f.c_path,f.c_modtime,f.c_mime,
		f.c_size,f.c_archiveId, a.c_path AS c_archivePath
		FROM t_page p 
		INNER JOIN t_file f ON p.c_fileId=f.c_id
		INNER JOIN t_archive a ON f.c_archiveId=a.c_id`)
	if err != nil {
		return nil, err
	}
	return pageRows(pgs).toModel(), nil
}

func (p Page) getPagesByIds(ctx context.Context, db dbtx, ids []model.ID) ([]model.Page, error) {
	var pgs []pageRow
	if len(ids) == 0 {
		return []model.Page{}, nil
	}
	q, args, err := sqlx.In(
		`SELECT p.c_id, p.c_width, p.c_height, p.c_isSpread, 
		p.c_fileId, p.c_pageName,f.c_path,f.c_modtime,f.c_mime,
		f.c_size,f.c_archiveId, a.c_path AS c_archivePath
		FROM t_page p 
		INNER JOIN t_file f ON p.c_fileId=f.c_id 
		INNER JOIN t_archive a ON f.c_archiveId=a.c_id
		WHERE p.c_id IN(?)`,
		ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)
	err = db.SelectContext(ctx, &pgs, q, args...)
	if err != nil {
		return nil, err
	}
	return pageRows(pgs).toModel(), nil
}

func (p Page) getPagesByFileIds(ctx context.Context, db dbtx, ids []model.ID) ([]model.Page, error) {
	var pgs []pageRow
	if len(ids) == 0 {
		return []model.Page{}, nil
	}
	q, args, err := sqlx.In(
		`SELECT p.c_id, p.c_width, p.c_height, p.c_isSpread, 
		p.c_fileId, p.c_pageName,f.c_path,f.c_modtime,f.c_mime,
		f.c_size,f.c_archiveId, a.c_path AS c_archivePath
		FROM t_page p 
		INNER JOIN t_file f ON p.c_fileId=f.c_id 
		INNER JOIN t_archive a ON f.c_archiveId=a.c_id
		WHERE p.c_fileId IN(?)`,
		ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)
	err = db.SelectContext(ctx, &pgs, q, args...)
	if err != nil {
		return nil, err
	}
	return pageRows(pgs).toModel(), nil
}
