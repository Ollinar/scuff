package db_test

import (
	"cmp"
	"context"
	"math/rand"
	"slices"
	"testing"

	"github.com/Ollinar/scuff/internal/model"

	"github.com/Ollinar/scuff/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

var seriesMockdata = []model.Series{
	{
		Title:      "dbz",
		Descripion: "chimp mode",
		ChapterIds: []int64{1, 2, 3, 4, 5},
	}, {
		Title:      "bleach",
		Descripion: "yummy",
		ChapterIds: []int64{6, 7, 8, 9, 10},
	},
}

func populateDBSeries(ctx context.Context) ([]model.Series, error) {
	s, err := db.NewSqlite(getDNS(true))
	if err != nil {
		return nil, err
	}
	defer s.Close()

	res := make([]model.Series, 0, len(seriesMockdata))
	for _, v := range seriesMockdata {
		srsTmp, err := s.Series().Add(ctx, v)
		if err != nil {
			return nil, err
		}
		res = append(res, srsTmp)

	}

	return res, nil
}

func cleanDBSeries(ctx context.Context) error {
	s, err := db.NewSqlite(getDNS(true))
	if err != nil {
		return err
	}
	defer s.Close()

	sl, err := s.Series().GetAll(ctx)
	if err != nil {
		return err
	}
	for _, v := range sl {
		err = s.Series().Remove(ctx, v.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestSeries_Add(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		series  model.Series
		want    model.Series
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "normal1",
			series: model.Series{
				Title:      "dbz",
				Descripion: "chimp mode",
				ChapterIds: []int64{1, 2, 3, 4, 5},
			},
			want: model.Series{
				Title:      "dbz",
				Descripion: "chimp mode",
				ChapterIds: []int64{1, 2, 3, 4, 5},
			},
		}, {
			name: "normal2",
			series: model.Series{
				Title:      "bleach",
				Descripion: "yum",
				ChapterIds: []int64{6, 7, 8, 9, 10},
			},
			want: model.Series{
				Title:      "bleach",
				Descripion: "yum",
				ChapterIds: []int64{6, 7, 8, 9, 10},
			},
		},
	}

	dbI, err := db.NewSqlite(getDNS(true))
	if err != nil {
		t.Fatal(err)
	}
	defer dbI.Close()

	s := dbI.Series()
	for tt := range slices.Values(tests) {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := s.Add(context.Background(), tt.series)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Add() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Add() succeeded unexpectedly")
			}
			if tt.want.Title != got.Title ||
				tt.want.Descripion != got.Descripion || !slices.Equal(tt.want.ChapterIds, got.ChapterIds) ||
				got.ID == 0 {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_GetAll(t *testing.T) {
	exp, err := populateDBSeries(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanDBSeries(t.Context())

	dbI, err := db.NewSqlite(getDNS(true))
	if err != nil {
		t.Fatal(err)
	}
	defer dbI.Close()
	s := dbI.Series()
	got, err := s.GetAll(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	slices.SortFunc(exp, func(a, b model.Series) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortFunc(got, func(a, b model.Series) int { return cmp.Compare(a.ID, b.ID) })
	if !slices.EqualFunc(exp, got, func(e1 model.Series, e2 model.Series) bool {
		return e1.ID == e2.ID && e1.Title == e2.Title && e1.Descripion == e2.Descripion && slices.Equal(e1.ChapterIds, e2.ChapterIds)
	}) {
		t.Errorf("mismatch of result, want: %v got: %v", exp, got)
	}
}

func TestSeries_Remove(t *testing.T) {
	res, err := populateDBSeries(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanDBSeries(t.Context())
	indx := rand.Intn(len(res))
	toBeRmvd := res[indx]

	dbI, err := db.NewSqlite(getDNS(true))
	if err != nil {
		t.Fatal(err)
	}
	defer dbI.Close()

	s := dbI.Series()
	err = s.Remove(t.Context(), toBeRmvd.ID)
	if err != nil {
		t.Fatal(err)
	}
	res2, err := s.GetAll(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range res {
		exst := slices.ContainsFunc(res2, func(e model.Series) bool { return e.ID == v.ID })
		if !exst {
			if v.ID != toBeRmvd.ID {
				t.Errorf("unexpected entry deleted %v", v)
			} else {
				t.Logf("succesfully deleted %v", v)
			}
		}
	}
}

func TestSeries_Update(t *testing.T) {
	res, err := populateDBSeries(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanDBSeries(t.Context())
	sqlite, err := db.NewSqlite(getDNS(true))
	if err != nil {
		t.Fatal(err)
	}

	sRepo := sqlite.Series()

	type updateFunc func(model.Series) model.Series
	type updateScenario struct {
		name string
		fn   updateFunc
	}

	updScen := []updateScenario{
		{
			name: "title change",
			fn: func(s model.Series) model.Series {
				s.Title = shuffleString(s.Title)
				return s
			},
		}, {
			name: "description change",
			fn: func(s model.Series) model.Series {
				s.Descripion = shuffleString(s.Descripion)
				return s
			},
		}, {
			name: "new chapterId",
			fn: func(s model.Series) model.Series {
				s.ChapterIds = addNewRandomInt(s.ChapterIds)
				return s
			},
		}, {
			name: "removed chapterId",
			fn: func(s model.Series) model.Series {
				s.ChapterIds = removeRandomElement(s.ChapterIds)
				return s
			},
		}, {
			name: "change chapterId order",
			fn: func(s model.Series) model.Series {
				s.ChapterIds = shuffleElements(s.ChapterIds)
				return s
			},
		},
	}

	updateCount := rand.Intn(1000)
	// updateCount := 1000

	for range updateCount {

		targetIdx := rand.Intn(len(res))
		toBeUpdated := res[targetIdx]

		numOfChange := rand.Intn(len(updScen))
		newSeries := toBeUpdated
		for range numOfChange {
			indxOfChange := rand.Intn(len(updScen))
			chng := updScen[indxOfChange]
			newSeries = chng.fn(toBeUpdated)
		}
		_, err = sRepo.Update(t.Context(), newSeries.ID, newSeries)
		if err != nil {
			t.Fatal(err)
		}
		resTmp, err := sRepo.GetByIDs(t.Context(), []model.ID{newSeries.ID})
		if err != nil {
			t.Fatal(err)
		}
		if len(resTmp) != 1 {
			t.Errorf("expected len of 1 result instead got: %v", resTmp)
		}
		upRes := resTmp[0]
		if newSeries.Title != upRes.Title || newSeries.Descripion != upRes.Descripion || !slices.Equal(newSeries.ChapterIds, upRes.ChapterIds) {
			t.Errorf("expected %v \n instead got :%v", newSeries, upRes)
		}

	}
}
