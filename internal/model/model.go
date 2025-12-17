// Package model defines the data models that the application uses.
package model

type ID = int64

type Archive struct {
	ID          ID
	Path        string
	Size        int64
	ModTime     int64
	Type        string
	PartialHash string
	FileIds     []ID
}

type File struct {
	ID          ID
	ArchiveID   ID
	ArchivePath string
	Path        string
	ModTime     int64
	Mime        string
	Size        int64
}

type Page struct {
	ID       ID
	Name     string
	Width    int
	Height   int
	IsSpread bool
	FileID   ID
	File     File
}

type Chapter struct {
	ID          ID
	Name        string
	Descripion  string
	CoverPageID ID
	PageIDs     []ID
	Tags        []Tag
}

type Tag struct {
	ID        ID
	Namespace string
	Label     string
}

type Series struct {
	ID         ID
	Title      string
	Descripion string
	ChapterIds []ID
}
