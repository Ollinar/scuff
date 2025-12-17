package blevesearcher_test

import (
	"testing"
	"time"

	"github.com/Ollinar/scuff/internal/blevesearcher"
	"github.com/Ollinar/scuff/internal/db"
	"github.com/Ollinar/scuff/internal/search"
)

func TestBelveSearcher_SearchChapter(t *testing.T) {
	bs, isnew, err := blevesearcher.NewBleveSeacher("./index.bleve")
	if err != nil {
		t.Fatal(err)
	}

	if isnew {
		sqlDB, err := db.NewSqlite("./../../test/app.db")
		if err != nil {
			t.Fatal(err)
		}
		defer sqlDB.Close()
		beginFetchChap := time.Now()
		chapters, err := sqlDB.Chapter().GetAll(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		afterFetch := time.Now()

		t.Logf("db fetch took: %vms", afterFetch.Sub(beginFetchChap).Milliseconds())
		// t.Log(len(chapters))
		err = bs.IndexChapters(t.Context(), chapters...)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf(" indexing took: %vms", time.Since(afterFetch).Milliseconds())

	}
	beginSearch := time.Now()
	limit := 2000
	page := 1
	ids, total, err := bs.SearchChapter(t.Context(), search.Pagination{
		PageSize: &limit,
		Page:     &page,
	}, &search.ChapterFilter{
		HaveTags: []search.TagFilter{
			{
				Namespace: search.StringFilter{Type: search.MatchingExact, Value: "artist"},
				Label:     search.StringFilter{Type: search.MatchingInfix, Value: "alp"},
			},
		},
	})
	// ids, err := bs.SearchChapter(t.Context(), search.Pagination{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("search took %vms", time.Since(beginSearch).Milliseconds())
	t.Logf("total hits: %d", total)
	t.Logf("%+v\n", ids)
}
