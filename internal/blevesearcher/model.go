package blevesearcher

import (
	"fmt"
	"strings"

	"github.com/Ollinar/scuff/internal/model"
)

type ChapterIndexModel struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	NameSortField string   `json:"nameSort"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
}

// BleveType implementes the search.BleveClassifier, this makes it so if a struct of this types gets indexed, it will use the mapping for chapter
// without this, the mapping for chapter wont be used for the indexing, breaking the index.
func (cim ChapterIndexModel) BleveType() string {
	return "chapter"
}

// NOTE: tags have to be denomralized, because it will get flatten when indexed if its and array of object.
// eg. it will become an array of namespace and an array of label, instead of an array of object with namespace and label
func denormalizeTags(tags []model.Tag) []string {
	tgs := make([]string, 0, len(tags))
	for _, v := range tags {
		tgs = append(tgs, fmt.Sprintf("%s:%s", v.Namespace, v.Label))
	}
	return tgs
}

func chapterIndexModelFromModels(chaps []model.Chapter) []ChapterIndexModel {
	chs := make([]ChapterIndexModel, len(chaps))
	for i, v := range chaps {
		chs[i].ID = v.ID
		chs[i].Name = v.Name
		chs[i].NameSortField = strings.ToLower(v.Name)
		chs[i].Description = v.Descripion
		chs[i].Tags = denormalizeTags(v.Tags)

	}
	return chs
}
