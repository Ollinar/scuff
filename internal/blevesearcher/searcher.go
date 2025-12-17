// Package blevesearcher implements the searcher innterface
package blevesearcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/repository"
	"github.com/Ollinar/scuff/internal/search"

	"github.com/blevesearch/bleve/v2"

	_ "github.com/blevesearch/bleve/v2/config"
	bleveSearch "github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
)

type BleveSearcher struct {
	mu           *sync.Mutex
	isRebuilding *atomic.Bool
	currentIndex bleve.Index
	indexPath    string
	index        bleve.IndexAlias
}

func NewBleveSeacher(indexPath string) (*BleveSearcher, bool, error) {
	newIndx := false
	index, err := bleve.Open(indexPath)
	if err != nil {
		if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
			newIndx = true
			index, err = buildIndex(indexPath)
			if err != nil {
				return nil, false, err
			}
		} else {
			return nil, false, err
		}
	}
	indexAl := bleve.NewIndexAlias(index)

	return &BleveSearcher{
		index: indexAl, indexPath: indexPath, currentIndex: index,
		isRebuilding: &atomic.Bool{}, mu: &sync.Mutex{},
	}, newIndx, nil
}

func (blvS BleveSearcher) Close() error {
	return blvS.index.Close()
}

func buildIndex(path string) (bleve.Index, error) {
	chapterMapping := bleve.NewDocumentMapping()

	// NOTE: we index the id too for sorting, sorting with docID doesnt behave expectedly for numeric id

	chapterIDField := bleve.NewNumericFieldMapping()
	chapterMapping.AddFieldMappingsAt("id", chapterIDField)

	chapterNameField := bleve.NewTextFieldMapping()
	chapterMapping.AddFieldMappingsAt("name", chapterNameField)
	chapterNameSortField := bleve.NewTextFieldMapping()
	chapterNameSortField.Analyzer = "keyword"
	chapterMapping.AddFieldMappingsAt("nameSort", chapterNameSortField)

	chapterDescriptionField := bleve.NewTextFieldMapping()
	chapterMapping.AddFieldMappingsAt("description", chapterDescriptionField)

	tagsField := bleve.NewTextFieldMapping()
	tagsField.Analyzer = "keyword"
	chapterMapping.AddFieldMappingsAt("tags", tagsField)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("chapter", chapterMapping)

	// index, err := bleve.NewMemOnly(indexMapping)
	index, err := bleve.New(path, indexMapping)
	if err != nil {
		return nil, err
	}
	return index, nil
}

func (blvS *BleveSearcher) RebuildIndex(ctx context.Context, repo repository.Chapter) error {
	if blvS.isRebuilding.Load() {
		return nil
	}
	blvS.isRebuilding.Swap(true)
	defer blvS.isRebuilding.Swap(false)

	tmpIndex := blvS.currentIndex
	// chack if the current index is temp from past failed rebuild
	// if not move the current index first
	if blvS.indexPath == blvS.currentIndex.Name() {
		// make sure tmp path is empty
		tmpPath := blvS.indexPath + ".tmp"
		err := os.RemoveAll(tmpPath)
		if err != nil {
			return err
		}
		err = os.CopyFS(tmpPath, os.DirFS(blvS.indexPath))
		if err != nil {
			return err
		}
		tmpIndex, err = bleve.Open(tmpPath)
		if err != nil {
			return err
		}
		oldIndex := blvS.currentIndex
		blvS.index.Swap([]bleve.Index{tmpIndex}, []bleve.Index{oldIndex})
		blvS.currentIndex = tmpIndex
		err = oldIndex.Close()
		if err != nil {
			return err
		}
	}
	// make sure the path is empty
	err := os.RemoveAll(blvS.indexPath)
	if err != nil {
		return err
	}

	newIndex, err := buildIndex(blvS.indexPath)
	if err != nil {
		return err
	}

	// TODO: make this maybe a streamable?
	chaps, err := repo.GetAll(ctx)
	if err != nil {
		newIndex.Close()
		return err
	}

	err = blvS.batchIndex(ctx, newIndex, chaps...)
	if err != nil {
		newIndex.Close()
		return err
	}

	blvS.index.Swap([]bleve.Index{newIndex}, []bleve.Index{tmpIndex})
	blvS.currentIndex = newIndex
	err = tmpIndex.Close()
	if err != nil {
		return err
	}

	return nil
}

func (blvS BleveSearcher) IndexChapters(ctx context.Context, chapters ...model.Chapter) error {
	return blvS.batchIndex(ctx, blvS.index, chapters...)
}

func (blvS BleveSearcher) batchIndex(ctx context.Context, index bleve.Index, chapters ...model.Chapter) error {
	batch := index.NewBatch()
	for _, chap := range chapterIndexModelFromModels(chapters) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if batch.Size() == 100 {
				err := index.Batch(batch)
				if err != nil {
					return err
				}
				batch = index.NewBatch()
			}
			err := batch.Index(fmt.Sprintf("%d", chap.ID), chap)
			if err != nil {
				return err
			}
		}
	}
	err := index.Batch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (blvS BleveSearcher) DeleteChapter(ctx context.Context, chapID model.ID) error {
	return blvS.index.Delete(fmt.Sprintf("%d", chapID))
}

func (blvS BleveSearcher) SearchChapter(ctx context.Context, pagination search.Pagination, filter *search.ChapterFilter) ([]model.ID, int, error) {
	index := blvS.index

	var mainQry query.Query = bleve.NewMatchAllQuery()
	if filter != nil {
		tagQuery := bleve.NewBooleanQuery()
		nameQuery := bleve.NewBooleanQuery()
		haveNameQuery := len(filter.HaveNames) > 0 || len(filter.NotHaveNames) > 0 || len(filter.ContainNames) > 0
		haveTagQuery := len(filter.HaveTags) > 0 || len(filter.NotHaveTags) > 0 || len(filter.ContainTags) > 0
		for _, v := range filter.HaveNames {
			rgx := bleve.NewRegexpQuery(fmt.Sprintf("(?i)%s", buildRegexpStringFilter(v)))
			rgx.SetField("name")
			nameQuery.AddMust(rgx)
		}
		for _, v := range filter.NotHaveNames {
			rgx := bleve.NewRegexpQuery(fmt.Sprintf("(?i)%s", buildRegexpStringFilter(v)))
			rgx.SetField("name")
			nameQuery.AddMustNot(rgx)
		}
		for _, v := range filter.ContainNames {
			rgx := bleve.NewRegexpQuery(fmt.Sprintf("(?i)%s", buildRegexpStringFilter(v)))
			rgx.SetField("name")
			nameQuery.AddShould(rgx)
		}
		for _, v := range filter.HaveTags {

			str := fmt.Sprintf("(?i)%s:%s",
				buildRegexpStringFilter(v.Namespace),
				buildRegexpStringFilter(v.Label))
			rgx := bleve.NewRegexpQuery(str)
			rgx.SetField("tags")
			tagQuery.AddMust(rgx)
		}
		for _, v := range filter.NotHaveTags {

			str := fmt.Sprintf("(?i)%s:%s",
				buildRegexpStringFilter(v.Namespace),
				buildRegexpStringFilter(v.Label))
			rgx := bleve.NewRegexpQuery(str)
			rgx.SetField("tags")
			tagQuery.AddMustNot(rgx)
		}
		for _, v := range filter.ContainTags {

			str := fmt.Sprintf("(?i)%s:%s",
				buildRegexpStringFilter(v.Namespace),
				buildRegexpStringFilter(v.Label))
			rgx := bleve.NewRegexpQuery(str)
			rgx.SetField("tags")
			tagQuery.AddShould(rgx)
		}

		newMainQuery := bleve.NewBooleanQuery()
		if haveNameQuery {
			newMainQuery.AddMust(nameQuery)
		}
		if haveTagQuery {
			newMainQuery.AddMust(tagQuery)
		}
		if haveNameQuery || haveTagQuery {
			mainQry = newMainQuery
		}
	}
	req := bleve.NewSearchRequest(mainQry)

	if pagination.PageSize != nil {
		req.Size = max(1, *pagination.PageSize)
	}
	if pagination.Page != nil {
		req.From = max(0, (*pagination.Page-1)*req.Size)
	}
	so := make(bleveSearch.SortOrder, 0, len(pagination.Sorting))
	for _, v := range pagination.Sorting {
		switch val := v.(type) {
		case search.IDSort:
			mode := bleveSearch.SortFieldMin
			if val.Direction() == search.Descending {
				mode = bleveSearch.SortFieldMax
			}
			so = append(so, &bleveSearch.SortField{
				Field:   "id",
				Desc:    val.Direction() == search.Descending,
				Type:    bleveSearch.SortFieldAsNumber,
				Mode:    mode,
				Missing: bleveSearch.SortFieldMissingLast,
			})
		case search.NameSort:
			mode := bleveSearch.SortFieldMin
			if val.Direction() == search.Descending {
				mode = bleveSearch.SortFieldMax
			}
			so = append(so, &bleveSearch.SortField{
				Field:   "nameSort",
				Desc:    val.Direction() == search.Descending,
				Type:    bleveSearch.SortFieldAsString,
				Mode:    mode,
				Missing: bleveSearch.SortFieldMissingLast,
			})
		case search.TagNamespaceSort:
			mode := bleveSearch.SortFieldMin
			if val.Direction() == search.Descending {
				mode = bleveSearch.SortFieldMax
			}
			so = append(so, &SortNamespace{
				Field:     "tags",
				Desc:      val.Direction() == search.Descending,
				Type:      bleveSearch.SortFieldAsString,
				Mode:      mode,
				Missing:   bleveSearch.SortFieldMissingLast,
				Namespace: val.Namespace,
			})
		}
	}

	if len(so) > 0 {
		req.Sort = so
	}

	res, err := index.SearchInContext(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]model.ID, 0, res.Total)
	for _, v := range res.Hits {
		id, err := strconv.Atoi(v.ID)
		if err != nil {
			return nil, 0, err
		}
		ids = append(ids, model.ID(id))
	}

	return ids, int(res.Total), nil
}

func buildRegexpStringFilter(f search.StringFilter) string {
	v := regexp.QuoteMeta(f.Value)
	switch f.Type {
	case search.MatchingExact:
		return v
	case search.MatchingInfix:
		return ".*" + v + ".*"
	case search.MatchingPrefix:
		return v + ".*"
	case search.MatchingSuffix:
		return ".*" + v
	default:
		return v
	}
}
