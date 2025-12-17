package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/search"
)

func (ap App) Chapter() chapterModule {
	return chapterModule{app: ap}
}

type chapterModule struct {
	app App
}

func (chm chapterModule) Add(ctx context.Context, chap model.Chapter) (model.Chapter, error) {
	if chap.Name == "" {
		return model.Chapter{}, errors.Join(ErrInvalidEntity, errors.New("chapter title is empty"))
	}

	// validate the PageIds
	if len(chap.PageIDs) > 0 {
		pgs, err := chm.app.pageRepo.GetByIDs(ctx, chap.PageIDs)
		if err != nil {
			return model.Chapter{}, errors.Join(ErrUnexpected, err)
		}
		pgsMap := make(map[model.ID]struct{}, len(pgs))
		for _, v := range pgs {
			pgsMap[v.ID] = struct{}{}
		}

		for _, v := range chap.PageIDs {
			if _, ok := pgsMap[v]; !ok {
				return model.Chapter{}, errors.Join(ErrInvalidEntity, fmt.Errorf("page with id of %d not found", v))
			}
		}
	}

	chap, err := chm.app.chapterRepo.Add(ctx, chap)
	if err != nil {
		return model.Chapter{}, errors.Join(ErrUnexpected, err)
	}
	err = chm.app.chapterSearcher.IndexChapters(ctx, chap)
	if err != nil {
		return model.Chapter{}, err
	}
	return chap, nil
}

func (chm chapterModule) GetAll(ctx context.Context) ([]model.Chapter, error) {
	chaps, err := chm.app.chapterRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return chaps, nil
}

func (chm chapterModule) GetByIDs(ctx context.Context, ids []model.ID) ([]model.Chapter, error) {
	chaps, err := chm.app.chapterRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	slices.SortStableFunc(chaps, func(a, b model.Chapter) int {
		return cmp.Compare(int(slices.Index(ids, a.ID)), int(slices.Index(ids, b.ID)))
	})
	return chaps, nil
}

func (chm chapterModule) Search(ctx context.Context, pagination search.Pagination, filter *search.ChapterFilter) ([]model.Chapter, int, error) {
	chapSearcher := chm.app.chapterSearcher
	chapIds, total, err := chapSearcher.SearchChapter(ctx, pagination, filter)
	if err != nil {
		return nil, 0, errors.Join(ErrUnexpected, err)
	}
	chaps, err := chm.GetByIDs(ctx, chapIds)
	if err != nil {
		return nil, 0, err
	}
	return chaps, total, nil
}

func (chm chapterModule) Update(ctx context.Context, chapterID model.ID, updates model.Chapter) (model.Chapter, error) {
	tmpChaps, err := chm.app.chapterRepo.GetByIDs(ctx, []model.ID{chapterID})
	if err != nil {
		return model.Chapter{}, errors.Join(ErrUnexpected, err)
	}
	if len(tmpChaps) < 1 {
		return model.Chapter{}, errors.Join(ErrNotFound, fmt.Errorf("failed to find chapter with id %d", chapterID))
	}
	if len(tmpChaps) > 1 {
		return model.Chapter{}, errors.Join(ErrUnexpected, fmt.Errorf("multiple chapter with id of %d found", chapterID))
	}

	curChap := tmpChaps[0]
	if updates.Name == "" {
		updates.Name = curChap.Name
	}

	tmpPgs, err := chm.app.pageRepo.GetByIDs(ctx, updates.PageIDs)
	if err != nil {
		return model.Chapter{}, errors.Join(ErrUnexpected, err)
	}
	tmpPgsMap := make(map[model.ID]struct{}, len(tmpPgs))
	for _, v := range tmpPgs {
		tmpPgsMap[v.ID] = struct{}{}
	}

	for _, v := range updates.PageIDs {
		if _, ok := tmpPgsMap[v]; !ok {
			return model.Chapter{}, errors.Join(ErrInvalidEntity, fmt.Errorf("page with id of %d not found", v))
		}
	}

	updates, err = chm.app.chapterRepo.Update(ctx, chapterID, updates)
	if err != nil {
		return model.Chapter{}, errors.Join(ErrUnexpected, err)
	}
	err = chm.app.chapterSearcher.IndexChapters(ctx, updates)
	if err != nil {
		return model.Chapter{}, err
	}

	return updates, nil
}

func (chm chapterModule) Remove(ctx context.Context, id model.ID) error {
	err := chm.app.chapterRepo.Remove(ctx, id)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	err = chm.app.chapterSearcher.DeleteChapter(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (chm chapterModule) AddChapterFromArchive(ctx context.Context, arcID model.ID) (model.Chapter, error) {
	arcs, err := chm.app.Archive().GetByIDs(ctx, []model.ID{arcID})
	if err != nil {
		return model.Chapter{}, err
	}
	if len(arcs) != 1 {
		return model.Chapter{}, errors.Join(ErrUnexpected, fmt.Errorf("expected to get 1 achive, insted got %d", len(arcs)))
	}
	arc := arcs[0]
	pgs, err := chm.app.Page().AddPagesFromArchive(ctx, arcID)
	if err != nil {
		return model.Chapter{}, err
	}
	chapName := filepath.Base(arc.Path)
	chapName = strings.TrimSuffix(chapName, filepath.Ext(chapName))

	pageIDs := make([]model.ID, len(pgs))
	for i, page := range pgs {
		pageIDs[i] = page.ID
	}

	chap := model.Chapter{
		Name:    chapName,
		PageIDs: pageIDs,
	}
	if len(pgs) > 0 {
		chap.CoverPageID = pgs[0].ID
	}

	chap, err = chm.Add(ctx, chap)
	if err != nil {
		return model.Chapter{}, err
	}
	// TODO: maybe add auto exec plugin for this
	return chap, nil
}
