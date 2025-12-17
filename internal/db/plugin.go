package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
)

type plugin struct {
	db *sqlx.DB
}

func newPlugin(db *sqlx.DB) plugin {
	return plugin{db: db}
}

func (pr plugin) StoreConfig(ctx context.Context, name string, ver string, conf map[string]string) error {
	confJ, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	_, err = pr.db.ExecContext(ctx,
		"INSERT INTO t_pluginConfig (c_name,c_version,c_configJson) VALUES(?,?,?) ON CONFLICT (c_name,c_version) DO UPDATE SET c_configJson=excluded.c_configJson",
		name, ver, string(confJ),
	)
	if err != nil {
		return err
	}
	return nil
}

func (pr plugin) GetConfig(ctx context.Context, name string, ver string) (map[string]string, error) {
	var confStr string
	err := pr.db.GetContext(ctx, &confStr, "SELECT p.c_configJson FROM t_pluginConfig p WHERE p.c_name=? AND p.c_version=?", name, ver)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var confMap map[string]string
	json.Unmarshal([]byte(confStr), &confMap)

	return confMap, nil
}
