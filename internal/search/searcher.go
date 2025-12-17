// Package search defines the common interface for serc functionalities used in application
package search

import (
	"context"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/repository"
)

type SortDirection int

const (
	Ascending = iota + 1
	Descending
)

type MatchingMode int

const (
	MatchingExact  = iota
	MatchingPrefix = iota
	MatchingSuffix = iota
	MatchingInfix  = iota
)

type Sort interface {
	Direction() SortDirection
}

type genericSort struct {
	direction SortDirection
}

func newGenericSort(direction SortDirection) genericSort {
	return genericSort{direction: direction}
}

type IDSort genericSort

func (is IDSort) Direction() SortDirection {
	return is.direction
}

func SortByID(direction SortDirection) IDSort {
	return IDSort(newGenericSort(direction))
}

type NameSort genericSort

func (ns NameSort) Direction() SortDirection {
	return ns.direction
}

func SortByName(direction SortDirection) NameSort {
	return NameSort(newGenericSort(direction))
}

type TagNamespaceSort struct {
	genericSort
	Namespace StringFilter
}

func (ts TagNamespaceSort) Direction() SortDirection {
	return ts.direction
}

func SortByTagNamespace(direction SortDirection, namespace StringFilter) TagNamespaceSort {
	return TagNamespaceSort{
		Namespace:   namespace,
		genericSort: newGenericSort(direction),
	}
}

type Pagination struct {
	// if absent, should default to all
	PageSize *int
	// 1 indexed, if absent, it means 1
	Page    *int
	Sorting []Sort
}

type StringFilter struct {
	Value string
	Type  MatchingMode
}

type TagFilter struct {
	Namespace StringFilter
	Label     StringFilter
}

type ChapterFilter struct {
	HaveNames    []StringFilter
	NotHaveNames []StringFilter
	ContainNames []StringFilter

	// chapter should have all of these tags.
	HaveTags []TagFilter
	// chapter shouldn't have all of these tags.
	NotHaveTags []TagFilter
	// chapter should have any of these tags, grouped by namespace.
	ContainTags []TagFilter
}

type ChapterSearcher interface {
	SearchChapter(ctx context.Context, pagination Pagination, filter *ChapterFilter) ([]model.ID, int, error)
	IndexChapters(ctx context.Context, chapters ...model.Chapter) error
	DeleteChapter(ctx context.Context, chapID model.ID) error
	RebuildIndex(ctx context.Context, repo repository.Chapter) error
}
