package api

import (
	"errors"
	"strings"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/search"
)

type ExecPluginJSON struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Target  int64             `json:"target"`
	Param   map[string]string `json:"param"`
}

type PluginSetConfigJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	AutoRun bool   `json:"autoRun"`
	Delay   int64  `json:"delay"`

	Config map[string]string `json:"config"`
}

type ChapterRequestJSON struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	Tags        []TagRequestJSON `json:"tags"`
	PageIDs     []int64          `json:"pageIds"`
	CoverPageID *int64           `json:"coverPageId"`
}

type TagRequestJSON struct {
	Namespace string `json:"namespace"`
	Label     string `json:"label"`
}

func (crj ChapterRequestJSON) parseTags() ([]model.Tag, error) {
	if crj.Tags == nil {
		return nil, nil
	}
	tags := make([]model.Tag, 0, len(crj.Tags))
	for _, tag := range crj.Tags {
		if tag.Label == "" {
			return nil, errors.New("tag label cant be empty")
		}
		if tag.Namespace == "" {
			tag.Namespace = "other"
		}
		tags = append(tags, model.Tag{Namespace: tag.Namespace, Label: tag.Label})
	}
	return tags, nil
}

type StringFilter struct {
	Value string `json:"value"`
	// Type could either be exact.suffix,prefix,infix
	Type string `json:"type"`
}

func (sf StringFilter) toSearchFilter() search.StringFilter {
	var ty search.MatchingMode
	switch strings.ToLower(sf.Type) {
	case "exact":
		ty = search.MatchingExact
	case "suffix":
		ty = search.MatchingSuffix
	case "prefix":
		ty = search.MatchingPrefix
	default:
		ty = search.MatchingInfix
	}

	return search.StringFilter{
		Type:  ty,
		Value: sf.Value,
	}
}

type TagFilter struct {
	Namespace StringFilter `json:"namespace"`
	Label     StringFilter `json:"label"`
}

func (tf TagFilter) toSearchTagFilter() search.TagFilter {
	return search.TagFilter{
		Namespace: tf.Namespace.toSearchFilter(),
		Label:     tf.Label.toSearchFilter(),
	}
}

type SortJSON struct {
	// by could either be id,name,tagNamespace
	By         string        `json:"by"`
	Descending bool          `json:"descending"`
	Namespace  *StringFilter `json:"namespace,omitempty"`
}

func (sj SortJSON) toSearchSort() search.Sort {
	var direction search.SortDirection = search.Ascending
	if sj.Descending {
		direction = search.Descending
	}
	switch strings.ToLower(sj.By) {
	case "id":
		return search.SortByID(direction)
	case "name":
		return search.SortByName(direction)
	case "tagnamespace":
		if sj.Namespace != nil {
			return search.SortByTagNamespace(direction, sj.Namespace.toSearchFilter())
		}
	default:

	}

	return search.SortByID(direction)
}

type ChapterFilterJSON struct {
	HaveNames    []StringFilter `json:"haveNames"`
	NotHaveNames []StringFilter `json:"notHaveNames"`
	ContainNames []StringFilter `json:"containNames"`

	// chapter should have all of these tags.
	HaveTags []TagFilter `json:"haveTags"`
	// chapter shouldn't have all of these tags.
	NotHaveTags []TagFilter `json:"notHaveTags"`
	// chapter should have any of these tags, grouped by namespace.
	ContainTags []TagFilter `json:"containTags"`

	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Sorting  []SortJSON `json:"sorting"`
}

func (cf ChapterFilterJSON) toFilterAndPagination() (*search.ChapterFilter, *search.Pagination) {
	pgnation := search.Pagination{
		Page:     &cf.Page,
		PageSize: &cf.PageSize,
	}
	for _, sfn := range cf.Sorting {
		pgnation.Sorting = append(pgnation.Sorting, sfn.toSearchSort())
	}
	if len(cf.HaveNames) == 0 && len(cf.NotHaveNames) == 0 && len(cf.ContainNames) == 0 && len(cf.HaveTags) == 0 && len(cf.NotHaveTags) == 0 && len(cf.ContainTags) == 0 {
		return nil, &pgnation
	}
	fltr := search.ChapterFilter{}

	for _, v := range cf.HaveNames {
		fltr.HaveNames = append(fltr.HaveNames, v.toSearchFilter())
	}
	for _, v := range cf.NotHaveNames {
		fltr.NotHaveNames = append(fltr.NotHaveNames, v.toSearchFilter())
	}
	for _, v := range cf.ContainNames {
		fltr.ContainNames = append(fltr.ContainNames, v.toSearchFilter())
	}

	for _, v := range cf.HaveTags {
		fltr.HaveTags = append(fltr.HaveTags, v.toSearchTagFilter())
	}
	for _, v := range cf.NotHaveTags {
		fltr.NotHaveTags = append(fltr.NotHaveTags, v.toSearchTagFilter())
	}
	for _, v := range cf.ContainTags {
		fltr.ContainTags = append(fltr.ContainTags, v.toSearchTagFilter())
	}
	return &fltr, &pgnation
}
