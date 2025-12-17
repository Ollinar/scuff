package app

import (
	"errors"
	"fmt"

	"github.com/Ollinar/scuff/internal/model"

	"golang.org/x/net/context"
)

func (ap App) SeriesModule() series {
	return series{app: ap}
}

type series struct {
	app App
}

func (sr series) AddSeries(ctx context.Context, series model.Series) (model.Series, error) {
	if series.Title == "" {
		return model.Series{}, errors.Join(ErrInvalidEntity, errors.New("series title is empty"))
	}

	// validate chapterIds
	if len(series.ChapterIds) > 0 {
		chaps, err := sr.app.chapterRepo.GetByIDs(ctx, series.ChapterIds)
		if err != nil {
			return model.Series{}, errors.Join(ErrUnexpected, err)
		}
		chapsMap := make(map[model.ID]struct{}, len(chaps))
		for _, v := range chaps {
			chapsMap[v.ID] = struct{}{}
		}
		for _, v := range series.ChapterIds {
			if _, ok := chapsMap[v]; !ok {
				return model.Series{}, errors.Join(ErrInvalidEntity, fmt.Errorf("chapter with id of %d not found", v))
			}
		}

	}
	series, err := sr.app.seriesRepo.Add(ctx, series)
	if err != nil {
		return model.Series{}, errors.Join(ErrUnexpected, err)
	}

	return series, nil
}

func (sr series) GetAllSeries(ctx context.Context) ([]model.Series, error) {
	s, err := sr.app.seriesRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return s, nil
}

func (sr series) GetSeriesByIds(ctx context.Context, ids []model.ID) ([]model.Series, error) {
	s, err := sr.app.seriesRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}
	return s, nil
}

func (sr series) UpdateSeries(ctx context.Context, seriesId model.ID, updates model.Series) (model.Series, error) {
	srsList, err := sr.app.seriesRepo.GetByIDs(ctx, []model.ID{seriesId})
	if err != nil {
		return model.Series{}, errors.Join(ErrUnexpected, err)
	}
	if len(srsList) < 1 {
		return model.Series{}, errors.Join(ErrNotFound, fmt.Errorf("failed to find series with id %d", seriesId))
	}
	if len(srsList) > 1 {
		return model.Series{}, errors.Join(ErrUnexpected, fmt.Errorf("multiple series with id of %d found", seriesId))
	}
	srs := srsList[0]

	// check to make sure title can never be empty
	if updates.Title == "" {
		updates.Title = srs.Title
	}

	chaps, err := sr.app.chapterRepo.GetByIDs(ctx, updates.ChapterIds)
	if err != nil {
		return model.Series{}, errors.Join(ErrUnexpected, err)
	}
	chapsMap := make(map[int64]struct{}, len(chaps))
	for _, v := range chaps {
		chapsMap[v.ID] = struct{}{}
	}
	for _, v := range updates.ChapterIds {
		if _, ok := chapsMap[v]; !ok {
			return model.Series{}, errors.Join(ErrInvalidEntity, fmt.Errorf("chapter with id of %d not found", v))
		}
	}

	updates, err = sr.app.seriesRepo.Update(ctx, seriesId, updates)
	if err != nil {
		return model.Series{}, errors.Join(ErrUnexpected, err)
	}

	return updates, nil
}

func (sr series) RemoveSeries(ctx context.Context, seriesId model.ID) error {
	err := sr.app.seriesRepo.Remove(ctx, seriesId)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	return nil
}
