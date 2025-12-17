// Package api defines and API using echo
package api

import (
	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/plugin"
)

type Archive struct {
	ID          int64   `json:"id"`
	Path        string  `json:"path"`
	Size        int64   `json:"size"`
	ModTime     int64   `json:"modtime"`
	Type        string  `json:"type"`
	PartialHash string  `json:"partialHash"`
	FileIDs     []int64 `json:"fileIDs"`
}

func (a Archive) toModel() model.Archive {
	return model.Archive{
		ID:          a.ID,
		Path:        a.Path,
		Size:        a.Size,
		ModTime:     a.ModTime,
		Type:        a.Type,
		PartialHash: a.PartialHash,
		FileIds:     a.FileIDs,
	}
}

func (a *Archive) fromModel(arc model.Archive) {
	a.ID = arc.ID
	a.Path = arc.Path
	a.Size = arc.Size
	a.ModTime = arc.ModTime
	a.Type = arc.Type
	a.PartialHash = arc.PartialHash
	a.FileIDs = arc.FileIds
}

func archivesFromModel(arcs []model.Archive) []Archive {
	al := make([]Archive, len(arcs))
	for i, v := range arcs {
		al[i].fromModel(v)
	}
	return al
}

type File struct {
	ID      int64  `json:"id"`
	Path    string `json:"path"`
	ModTime int64  `json:"modifiedAt"`
	Mime    string `json:"mime"`
	Size    int64  `json:"Size"`
}

func (f *File) fromModel(fl model.File) {
	f.ID = fl.ID
	f.Path = fl.Path
	f.ModTime = fl.ModTime
	f.Mime = fl.Mime
	f.Size = fl.Size
}

func filesFromModel(fls []model.File) []File {
	fL := make([]File, len(fls))
	for i, fl := range fls {
		fL[i].fromModel(fl)
	}
	return fL
}

type Page struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	IsSpread bool   `json:"isSpread"`
	FileID   int64  `json:"fileID"`
	File     File   `json:"file"`
}

func (pg *Page) fromModel(page model.Page) {
	fl := File{}
	fl.fromModel(page.File)
	pg.File = fl
	pg.ID = page.ID
	pg.Name = page.Name
	pg.Width = page.Width
	pg.Height = page.Height
	pg.IsSpread = page.IsSpread
	pg.FileID = page.FileID
}

func pagesFromModel(pgs []model.Page) []Page {
	pages := make([]Page, len(pgs))

	for i, page := range pgs {
		pages[i].fromModel(page)
	}
	return pages
}

type Tag struct {
	Namespace string `json:"namespace"`
	Label     string `json:"label"`
}

func (t *Tag) fromModel(tag model.Tag) {
	t.Namespace = tag.Namespace
	t.Label = tag.Label
}

func tagsFromModel(tags []model.Tag) []Tag {
	tgs := make([]Tag, len(tags))
	for i, tag := range tags {
		tgs[i].fromModel(tag)
	}
	return tgs
}

type Chapter struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	PageIDs     []int64 `json:"pageIDs"`
	CoverPageID int64   `json:"coverPageId"`
	Tags        []Tag   `json:"tags"`
}

func (ch *Chapter) fromModel(chapter model.Chapter) {
	ch.ID = chapter.ID
	ch.Name = chapter.Name
	ch.Description = chapter.Descripion
	ch.PageIDs = chapter.PageIDs
	ch.Tags = tagsFromModel(chapter.Tags)
	ch.CoverPageID = chapter.CoverPageID
	if ch.CoverPageID == 0 && len(ch.PageIDs) > 0 {
		ch.CoverPageID = ch.PageIDs[0]
	}
}

func chaptersFromModel(chapters []model.Chapter) []Chapter {
	chaps := make([]Chapter, len(chapters))
	for i, chapter := range chapters {
		chaps[i].fromModel(chapter)
	}
	return chaps
}

type PluginConfig struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (pc *PluginConfig) fromModel(pic plugin.PluginConfig) {
	pc.Name = pic.Name
	pc.Value = pic.Value
	pc.Description = pic.Description
}

func pluginConfigsFromModel(plugConfs []plugin.PluginConfig) []PluginConfig {
	pcs := make([]PluginConfig, len(plugConfs))
	for i, conf := range plugConfs {
		pcs[i].fromModel(conf)
	}
	return pcs
}

type PluginParam = PluginConfig

type PluginInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	// Delay in milliseconds when executing multiple times
	Delay   int64 `json:"delay"`
	AutoRun bool  `json:"autoRun"`
	// TargetEntity specifies what the plugin should be ran against
	TargetEntity string         `json:"targetType"`
	Config       []PluginConfig `json:"config"`
	Param        []PluginParam  `json:"param"`
}

func (pi *PluginInfo) fromModel(plugInfs plugin.PluginInfo) {
	pi.Name = plugInfs.Name
	pi.Description = plugInfs.Description
	pi.Version = plugInfs.Version
	pi.Delay = plugInfs.Delay.Milliseconds()
	pi.AutoRun = plugInfs.AutoRun
	pi.Config = pluginConfigsFromModel(plugInfs.Config)
	pi.Param = pluginConfigsFromModel(plugInfs.Param)
	pi.TargetEntity = string(plugInfs.TargetEntity)
}

func pluginInfosFromModel(plugInfs []plugin.PluginInfo) []PluginInfo {
	pfs := make([]PluginInfo, len(plugInfs))
	for i, pi := range plugInfs {
		pfs[i].fromModel(pi)
	}
	return pfs
}
