package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/Ollinar/scuff/internal/model"

	"github.com/jmoiron/sqlx"
)

type Chapter struct {
	db *sqlx.DB
}

func newChapter(db *sqlx.DB) Chapter {
	return Chapter{db: db}
}

// Add implements repository.Chapter.
func (ch Chapter) Add(ctx context.Context, chap model.Chapter) (model.Chapter, error) {
	tx, err := ch.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.Chapter{}, err
	}
	defer rollbackTxx(tx)

	chaps, err := ch.addChapters(ctx, tx, []model.Chapter{chap})
	if err != nil {
		return model.Chapter{}, err
	}
	if len(chaps) != 1 {
		return model.Chapter{}, fmt.Errorf("expected 1 chapter to be added, insted got %d", len(chaps))
	}
	chap = chaps[0]

	err = tx.Commit()
	if err != nil {
		return model.Chapter{}, err
	}

	return chap, nil
}

// GetAll implements repository.Chapter.
func (ch Chapter) GetAll(ctx context.Context) ([]model.Chapter, error) {
	return ch.getAllChapters(ctx, ch.db)
}

// GetByIDs implements repository.Chapter.
func (ch Chapter) GetByIDs(ctx context.Context, ids []model.ID) ([]model.Chapter, error) {
	return ch.getChaptersByIds(ctx, ch.db, ids)
}

// Remove implements repository.Chapter.
func (ch Chapter) Remove(ctx context.Context, id model.ID) error {
	tx, err := ch.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackTxx(tx)

	err = ch.removeChapters(ctx, tx, []model.ID{id})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Update implements repository.Chapter.
func (ch Chapter) Update(ctx context.Context, chapterID model.ID, chap model.Chapter) (model.Chapter, error) {
	tx, err := ch.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.Chapter{}, err
	}
	defer rollbackTxx(tx)
	err = ch.update(ctx, tx, chapterID, chap)
	if err != nil {
		return model.Chapter{}, err
	}

	err = tx.Commit()
	if err != nil {
		return model.Chapter{}, err
	}
	return chap, nil
}

// func (ch Chapter) Search(ctx context.Context, pagination search.Pagination, filter *search.ChapterFilter) ([]model.Chapter, int, error) {
// 	return ch.search(ctx, ch.db, pagination, filter)
// }

func (ch Chapter) addChapters(ctx context.Context, tx *sqlx.Tx, chaps []model.Chapter) ([]model.Chapter, error) {
	if len(chaps) == 0 {
		return []model.Chapter{}, nil
	}

	insChapStm, err := tx.PreparexContext(ctx, "INSERT INTO t_chapter (c_name,c_description,c_coverPageId) VALUES (?,?,?)")
	if err != nil {
		return nil, err
	}
	defer insChapStm.Close()

	var coverPageID sql.NullInt64
	for chapIdx, chap := range chaps {
		coverPageID.Valid = chap.CoverPageID != 0
		coverPageID.Int64 = chap.CoverPageID
		res, err := insChapStm.ExecContext(ctx, chap.Name, chap.Descripion, coverPageID)
		if err != nil {
			return nil, err
		}
		chap.ID, err = res.LastInsertId()
		if err != nil {
			return nil, err
		}
		if len(chap.PageIDs) != 0 {
			insChapPageStm, err := tx.PreparexContext(ctx, "INSERT INTO t_chapterPage (c_chapterId,c_pageId,c_pageNumber) VALUES (?,?,?)")
			if err != nil {
				return nil, err
			}
			defer insChapPageStm.Close()
			for i, pageID := range chap.PageIDs {
				_, err = insChapPageStm.ExecContext(ctx, chap.ID, pageID, i+1)
				if err != nil {
					return nil, err
				}
			}
		}

		if len(chap.Tags) != 0 {
			tagRws, err := ch.getsertTags(ctx, tx, chap.Tags)
			if err != nil {
				return nil, err
			}

			tagIds := make([]model.ID, 0, len(tagRws))
			for _, tr := range tagRws {
				tagIds = append(tagIds, tr.CID)
			}
			err = ch.addChapterTags(ctx, tx, chap.ID, tagIds)
			if err != nil {
				return nil, err
			}

			chap.Tags = tagRows(tagRws).toModel()
		}

		chaps[chapIdx] = chap
	}

	return chaps, nil
}

func (ch Chapter) update(ctx context.Context, tx *sqlx.Tx, chapterID model.ID, chap model.Chapter) error {
	currChaps, err := ch.getChaptersByIds(ctx, tx, []int64{chapterID})
	if err != nil {
		return err
	}
	if len(currChaps) != 1 {
		return errors.New("unexpected number of chapter found")
	}
	currChap := currChaps[0]

	var coverPageID *int64 = nil
	if chap.CoverPageID > 0 {
		coverPageID = &chap.CoverPageID
	}

	_, err = tx.ExecContext(ctx, "UPDATE t_chapter SET c_name=?,c_description=?,c_coverPageId=? WHERE c_id=?",
		chap.Name, chap.Descripion, coverPageID, chapterID)
	if err != nil {
		return err
	}

	tags, err := ch.getsertTags(ctx, tx, chap.Tags)
	if err != nil {
		return err
	}

	removedTagIds := make([]model.ID, 0, len(currChap.Tags))
	for _, tg := range currChap.Tags {
		isRemoved := !slices.ContainsFunc(tags, func(tr tagRow) bool { return tg.Namespace == tr.CNamespace && tg.Label == tr.CLabel })
		if isRemoved {
			removedTagIds = append(removedTagIds, tg.ID)
		}
	}

	newTagIds := make([]model.ID, 0, len(tags))
	for _, tr := range tags {
		isNew := !slices.ContainsFunc(currChap.Tags, func(tg model.Tag) bool { return tg.Namespace == tr.CNamespace && tg.Label == tr.CLabel })
		if isNew {
			newTagIds = append(newTagIds, tr.CID)
		}
	}

	if len(removedTagIds) > 0 {
		err = ch.removeChapterTags(ctx, tx, chapterID, removedTagIds)
		if err != nil {
			return err
		}
	}

	if len(newTagIds) > 0 {
		err = ch.addChapterTags(ctx, tx, chapterID, newTagIds)
		if err != nil {
			return err
		}
	}

	for pageIndex, pageID := range diffIdsSeq(chap.PageIDs, currChap.PageIDs, false) {
		// TODO: come up with better way of upserting changes
		err = ch.upsertChapterPage(ctx, tx, chapterID, pageID, pageIndex+1)
		if err != nil {
			return err
		}
	}
	// cleanup trailing page grater than pages length
	_, err = tx.ExecContext(ctx, "DELETE FROM t_chapterPage WHERE c_chapterId =? AND c_pageNumber>?", chapterID, len(chap.PageIDs))
	if err != nil {
		return err
	}

	return nil
}

func (ch Chapter) removeChapters(ctx context.Context, tx *sqlx.Tx, chapIds []model.ID) error {
	if len(chapIds) == 0 {
		return nil
	}
	q, args, err := sqlx.In("DELETE FROM t_chapter WHERE c_id IN(?)", chapIds)
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

func (ch Chapter) getAllChapters(ctx context.Context, db dbtx) ([]model.Chapter, error) {
	var chaps []chapterRow

	// query has to be left join and group by so chapter with no pages also gets included and coalesce incase group_concat returns a null
	// alias is so sqlx could parse it
	err := db.SelectContext(ctx, &chaps,
		`SELECT c.c_id,c.c_name,c.c_description,c.c_coverPageId, COALESCE(GROUP_CONCAT(cp.c_pageId ORDER BY cp.c_pageNumber), "") AS pageIds
	FROM t_chapter c
	LEFT JOIN t_chapterPage cp ON c.c_id=cp.c_chapterId
	GROUP BY c.c_id`)
	if err != nil {
		return nil, err
	}

	for i, v := range chaps {
		tgs, err := ch.getChapterTags(ctx, db, v.CID)
		if err != nil {
			return nil, err
		}
		chaps[i].tags = tgs

	}

	return chapterRows(chaps).toModel(), nil
}

func (ch Chapter) getChaptersByIds(ctx context.Context, db dbtx, ids []int64) ([]model.Chapter, error) {
	var chaps []chapterRow
	if len(ids) < 1 {
		return []model.Chapter{}, nil
	}

	// query has to be left join and group by so chapter with no pages also gets included and coalesce incase group_concat returns a null
	// alias is so sqlx could parse it
	q, args, err := sqlx.In(`SELECT c.c_id,c.c_name,c.c_description,c.c_coverPageId,
		COALESCE(GROUP_CONCAT(cp.c_pageId ORDER BY cp.c_pageNumber),"") AS pageIds
	FROM t_chapter c
	LEFT JOIN t_chapterPage cp ON c.c_id=cp.c_chapterId
	WHERE c.c_id IN (?)
	GROUP BY c.c_id`, ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)

	err = db.SelectContext(ctx, &chaps, q, args...)
	if err != nil {
		return nil, err
	}

	for i, v := range chaps {
		tgs, err := ch.getChapterTags(ctx, db, v.CID)
		if err != nil {
			return nil, err
		}
		chaps[i].tags = tgs

	}

	return chapterRows(chaps).toModel(), nil
}

func (ch Chapter) upsertChapterPage(ctx context.Context, tx *sqlx.Tx, chapterID, pageID model.ID, pageNumber int) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO t_chapterPage (c_chapterId,c_pageId,c_pageNumber) VALUES (?,?,?)
				ON CONFLICT (c_chapterId,c_pageNumber) DO UPDATE SET c_pageId = excluded.c_pageId`,
		chapterID, pageID, pageNumber)
	return err
}

func (ch Chapter) addChapterTags(ctx context.Context, tx *sqlx.Tx, chapterID model.ID, tagIds []model.ID) error {
	if len(tagIds) == 0 {
		return nil
	}
	stm, err := tx.PreparexContext(ctx,
		"INSERT INTO t_chapterTag (c_chapterId,c_tagId) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer stm.Close()
	for _, tagID := range tagIds {
		_, err = stm.ExecContext(ctx,
			chapterID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch Chapter) removeChapterTags(ctx context.Context, tx *sqlx.Tx, chapterID model.ID, tagIds []model.ID) error {
	if len(tagIds) == 0 {
		return nil
	}
	q, args, err := sqlx.In("DELETE FROM t_chapterTag WHERE c_chapterId=? AND c_tagId IN(?)",
		chapterID, tagIds)
	if err != nil {
		return err
	}
	q = tx.Rebind(q)

	_, err = tx.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return err
}

func (ch Chapter) getChapterTags(ctx context.Context, db dbtx, chapterID int64) ([]model.Tag, error) {
	var tgs []tagRow

	err := db.SelectContext(ctx, &tgs, `SELECT t.c_id,t.c_namespace,t.c_label FROM t_tag t
		INNER JOIN t_chapterTag tc ON t.c_id=tc.c_tagId
		WHERE tc.c_chapterId = ?`,
		chapterID)
	if err != nil {
		return nil, err
	}
	return tagRows(tgs).toModel(), nil
}

// getsertTag will fetch or insert a tag and return the id
func (ch Chapter) getsertTags(ctx context.Context, tx *sqlx.Tx, tags []model.Tag) ([]tagRow, error) {
	if len(tags) == 0 {
		return []tagRow{}, nil
	}

	// using concatinated values that way we can use IN clause
	tagConcats := make([]string, 0, len(tags))
	for _, v := range tags {
		tagConcats = append(tagConcats, v.Namespace+v.Label)
	}

	q, args, err := sqlx.In("SELECT t.c_id,t.c_namespace,t.c_label FROM t_tag t WHERE CONCAT(t.c_namespace,t.c_label) IN (?)", tagConcats)
	if err != nil {
		return nil, err
	}
	q = tx.Rebind(q)
	var existingTags []tagRow
	err = tx.SelectContext(ctx, &existingTags, q, args...)
	if err != nil {
		return nil, err
	}

	missingTags := make([]tagRow, 0, len(tags))
	for _, v := range tags {
		ok := slices.ContainsFunc(existingTags, func(e tagRow) bool { return v.Label == e.CLabel && v.Namespace == e.CNamespace })
		if ok {
			continue
		}
		missingTags = append(missingTags, tagRow{
			CNamespace: v.Namespace,
			CLabel:     v.Label,
		})
	}

	if len(missingTags) > 0 {
		stm, err := tx.PreparexContext(ctx, "INSERT INTO t_tag (c_namespace,c_label) VALUES (?,?)")
		if err != nil {
			return nil, err
		}
		defer stm.Close()
		for _, v := range missingTags {
			res, err := stm.ExecContext(ctx, v.CNamespace, v.CLabel)
			if err != nil {
				return nil, err
			}
			v.CID, err = res.LastInsertId()
			if err != nil {
				return nil, err
			}
			existingTags = append(existingTags, v)
		}
	}

	return existingTags, nil
}
