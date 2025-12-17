package db

import (
	"context"
	"errors"

	"github.com/Ollinar/scuff/internal/model"

	"github.com/jmoiron/sqlx"
)

type Series struct {
	db *sqlx.DB
}

// Add implements repository.Series.
func (s Series) Add(ctx context.Context, series model.Series) (model.Series, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.Series{}, err
	}
	defer rollbackTxx(tx)
	res, err := tx.ExecContext(ctx, "INSERT INTO t_series (c_title,c_description) VALUES (?,?)", series.Title, series.Descripion)
	if err != nil {
		return model.Series{}, err
	}
	seriesId, err := res.LastInsertId()
	if err != nil {
		return model.Series{}, err
	}
	series.ID = seriesId
	for chapterIndex, chapterId := range series.ChapterIds {
		err = s.upsertSeriesChapter(ctx, tx, seriesId, chapterId, chapterIndex+1)
		if err != nil {
			return model.Series{}, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return model.Series{}, err
	}
	return series, nil
}

// GetAll implements repository.Series.
func (s Series) GetAll(ctx context.Context) ([]model.Series, error) {
	var sl []seriesRow
	// query has to be left join and group by so series with no chapter also gets included and coalesce incase group_concat returns a null
	// alias is so sqlx could parse it
	err := s.db.SelectContext(ctx, &sl,
		`SELECT s.c_id,s.c_title,c_description, COALESCE(GROUP_CONCAT(sc.c_chapterId),"") AS chapterIds 
		FROM t_series s 
		LEFT JOIN t_seriesChapter sc ON s.c_id = sc.c_seriesId 
		GROUP BY s.c_id`)
	if err != nil {
		return nil, err
	}
	return seriesRows(sl).toModel(), nil
}

// GetByIDs implements repository.Series.
func (s Series) GetByIDs(ctx context.Context, ids []model.ID) ([]model.Series, error) {
	return s.getByIDs(ctx, s.db, ids)
}

func (s Series) getByIDs(ctx context.Context, db dbtx, ids []model.ID) ([]model.Series, error) {
	sl := []seriesRow{}
	if len(ids) < 1 {
		return []model.Series{}, nil
	}

	// query has to be left join and group by so series with no chapter also gets included and coalesce incase group_concat returns a null
	// alias is so sqlx could parse it
	q, args, err := sqlx.In(
		`SELECT s.c_id,s.c_title,c_description, COALESCE(GROUP_CONCAT(sc.c_chapterId),"") AS chapterIds 
		FROM t_series s 
		LEFT JOIN t_seriesChapter sc ON s.c_id = sc.c_seriesId 
		WHERE s.c_id IN (?) 
		GROUP BY s.c_id`, ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)
	err = db.SelectContext(ctx, &sl, q, args...)
	if err != nil {
		return nil, err
	}
	return seriesRows(sl).toModel(), nil
}

// Remove implements repository.Series.
func (s Series) Remove(ctx context.Context, id model.ID) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM t_series WHERE c_id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

// Update implements repository.Series.
func (s Series) Update(ctx context.Context, id model.ID, series model.Series) (model.Series, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.Series{}, err
	}
	defer rollbackTxx(tx)
	currTmp, err := s.getByIDs(ctx, tx, []model.ID{id})
	if err != nil {
		return model.Series{}, err
	}
	if len(currTmp) != 1 {
		return model.Series{}, errors.New("unexpected number of series found")
	}
	currS := currTmp[0]

	_, err = tx.ExecContext(ctx, "UPDATE t_series SET c_title=?,c_description=? WHERE c_id=?", series.Title, series.Descripion, id)
	if err != nil {
		return model.Series{}, err
	}

	// purpose of diffIdsSeq is so operations are only done to those that need it.
	for chapterIndex, chapterId := range diffIdsSeq(series.ChapterIds, currS.ChapterIds, false) {
		err = s.upsertSeriesChapter(ctx, tx, id, chapterId, chapterIndex+1)
		if err != nil {
			return model.Series{}, err
		}
	}
	// cleanup trailing chapNum beyond the last one
	_, err = tx.ExecContext(ctx, "DELETE FROM t_seriesChapter WHERE c_seriesId =? AND c_chapterNumber>?", id, len(series.ChapterIds))
	if err != nil {
		return model.Series{}, err
	}

	err = tx.Commit()
	if err != nil {
		return model.Series{}, err
	}
	return series, nil
}

func (sr Series) upsertSeriesChapter(ctx context.Context, tx dbtx, seriesId, chapterId model.ID, chapterNumber int) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO t_seriesChapter (c_seriesId,c_chapterId,c_chapterNumber) 
					VALUES (?,?,?) ON CONFLICT(c_seriesId,c_chapterNumber) DO UPDATE SET c_chapterId = excluded.c_chapterId`,
		seriesId, chapterId, chapterNumber,
	)
	return err
}

func newSeries(db *sqlx.DB) Series {
	return Series{db: db}
}
