// Package db implrements the repository pacakge
package db

import (
	"github.com/Ollinar/scuff/internal/model"
)

type fileRow struct {
	CID          int64  `db:"c_id"`
	CPath        string `db:"c_path"`
	CModTime     int64  `db:"c_modtime"`
	CMime        string `db:"c_mime"`
	CSize        int64  `db:"c_size"`
	CArchiveID   int64  `db:"c_archiveId"`
	CArchivePath string `db:"c_archivePath"`
}

func (fl fileRow) toModel() model.File {
	return model.File{
		ID:          fl.CID,
		Path:        fl.CPath,
		Size:        fl.CSize,
		ModTime:     fl.CModTime,
		Mime:        fl.CMime,
		ArchiveID:   fl.CArchiveID,
		ArchivePath: fl.CArchivePath,
	}
}

type filerows []fileRow

func (fr filerows) toModel() []model.File {
	fls := make([]model.File, 0, len(fr))
	for _, v := range fr {
		fls = append(fls, v.toModel())
	}
	return fls
}

type archiveRow struct {
	CID          int64  `db:"c_id"`
	CPath        string `db:"c_path"`
	CSize        int64  `db:"c_size"`
	CModTime     int64  `db:"c_modtime"`
	CType        string `db:"c_type"`
	CPartialHash string `db:"c_partialhash"`
	CFileIDs     string `db:"fileIds"`
}

func (at archiveRow) toModel() model.Archive {
	fileIds, err := parseIdsStr(at.CFileIDs)
	if err != nil {
		panic("failed to parse page ids ")
	}

	return model.Archive{
		ID:          at.CID,
		Path:        at.CPath,
		Size:        at.CSize,
		ModTime:     at.CModTime,
		Type:        at.CType,
		FileIds:     fileIds,
		PartialHash: at.CPartialHash,
	}
}

type archiveRows []archiveRow

func (ar archiveRows) toModel() []model.Archive {
	arcs := make([]model.Archive, 0, len(ar))
	for _, arc := range ar {
		arcs = append(arcs, arc.toModel())
	}
	return arcs
}

type pageRow struct {
	CID       int64  `db:"c_id"`
	CWidth    int    `db:"c_width"`
	CHeight   int    `db:"c_height"`
	CIsSpread bool   `db:"c_isSpread"`
	CFileID   int64  `db:"c_fileId"`
	CPageName string `db:"c_pageName"`
	fileRow
}

func (pr pageRow) toModel() model.Page {
	fl := model.Page{
		ID:       pr.CID,
		Name:     pr.CPageName,
		Width:    pr.CWidth,
		Height:   pr.CHeight,
		IsSpread: pr.CIsSpread,
		FileID:   pr.CFileID,
		File:     pr.fileRow.toModel(),
	}
	fl.File.ID = pr.CFileID
	return fl
}

type pageRows []pageRow

func (pr pageRows) toModel() []model.Page {
	pL := make([]model.Page, 0, len(pr))
	for _, v := range pr {
		pL = append(pL, v.toModel())
	}
	return pL
}

type tagRow struct {
	CID        int64  `db:"c_id"`
	CNamespace string `db:"c_namespace"`
	CLabel     string `db:"c_label"`
}

func (t tagRow) toModel() model.Tag {
	return model.Tag{
		ID:        t.CID,
		Namespace: t.CNamespace,
		Label:     t.CLabel,
	}
}

type tagRows []tagRow

func (tl tagRows) toModel() []model.Tag {
	tgLs := make([]model.Tag, 0, len(tl))
	for _, v := range tl {
		tgLs = append(tgLs, v.toModel())
	}
	return tgLs
}

type chapterRow struct {
	CID          int64  `db:"c_id"`
	CName        string `db:"c_name"`
	CDescription string `db:"c_description"`
	tags         []model.Tag
	CCoverPageID *int64 `db:"c_coverPageId"`
	PageIds      string `db:"pageIds"`
	Count        int64  `db:"count"`
}

func (cr chapterRow) toModel() model.Chapter {
	ids, err := parseIdsStr(cr.PageIds)
	if err != nil {
		panic("failed to parse page ids ")
	}
	var coverID model.ID
	if cr.CCoverPageID != nil {
		coverID = model.ID(*cr.CCoverPageID)
	}
	return model.Chapter{
		ID:          cr.CID,
		Name:        cr.CName,
		Descripion:  cr.CDescription,
		PageIDs:     ids,
		CoverPageID: coverID,
		Tags:        cr.tags,
	}
}

type chapterRows []chapterRow

func (cr chapterRows) toModel() []model.Chapter {
	crs := make([]model.Chapter, 0, len(cr))
	for _, v := range cr {
		crs = append(crs, v.toModel())
	}
	return crs
}

type seriesRow struct {
	CID          int64  `db:"c_id"`
	CTitle       string `db:"c_title"`
	CDescription string `db:"c_description"`
	ChapterIDs   string `db:"chapterIds"`
}

func (srs seriesRow) toModel() model.Series {
	ids, err := parseIdsStr(srs.ChapterIDs)
	if err != nil {
		panic("failed to parse chapter ids ")
	}
	return model.Series{
		ID:         srs.CID,
		Title:      srs.CTitle,
		Descripion: srs.CDescription,
		ChapterIds: ids,
	}
}

type seriesRows []seriesRow

func (sl seriesRows) toModel() []model.Series {
	srs := make([]model.Series, 0, len(sl))
	for _, v := range sl {
		srs = append(srs, v.toModel())
	}
	return srs
}
