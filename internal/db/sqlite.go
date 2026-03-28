package db

import (
	"database/sql"
	"embed"
	"errors"
	"net/url"
	"strings"

	"github.com/Ollinar/scuff/internal/repository"

	"github.com/golang-migrate/migrate/v4"
	migrate_sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

//go:embed schema/*.sql
var schemaFs embed.FS

type Sqlite struct {
	db          *sqlx.DB
	archiveRepo repository.Archive
	pageRepo    repository.Page
	chapterRepo repository.Chapter
	seriesRepo  repository.Series
	pluginRepo  repository.Plugin
}

func NewSqlite(dsn string) (*Sqlite, error) {

	mDB, err := sqlx.Open("sqlite3", addOptionsToDSN(dsn))
	if err != nil {
		return nil, err
	}
	err = migrateDB(mDB.DB)
	if err != nil {
		return nil, err
	}
	sq := &Sqlite{
		db:          mDB,
		archiveRepo: newArchive(mDB),
		pageRepo:    newPage(mDB),
		chapterRepo: newChapter(mDB),
		seriesRepo:  newSeries(mDB),
		pluginRepo:  newPlugin(mDB),
	}
	return sq, nil
}

func migrateDB(db *sql.DB) error {
	// NOTE: we only close the source driver, since closing the migrate or the driver wrapper also closes db?
	sd, err := iofs.New(schemaFs, "schema")
	if err != nil {
		return err
	}
	defer sd.Close()

	sqlDB, err := migrate_sqlite3.WithInstance(db,
		&migrate_sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("schema migration",
		sd, "sqlite",
		sqlDB,
	)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// addOptionsToDSN add necessary options to dns
func addOptionsToDSN(dns string) string {
	srcStr, currOpts, _ := strings.Cut(dns, "?")
	vals, err := url.ParseQuery(currOpts)
	if err != nil {
		vals = make(url.Values)
	}
	vals.Set("_fk", "true")
	vals.Set("_journal", "wal")

	return srcStr + "?" + vals.Encode()
}

func (s Sqlite) Close() error {
	return s.db.Close()
}

func (s Sqlite) Archive() repository.Archive {
	return s.archiveRepo
}

func (s Sqlite) Page() repository.Page {
	return s.pageRepo
}

func (s Sqlite) Chapter() repository.Chapter {
	return s.chapterRepo
}

func (s Sqlite) Series() repository.Series {
	return s.seriesRepo
}

func (s Sqlite) Plugin() repository.Plugin {
	return s.pluginRepo
}
