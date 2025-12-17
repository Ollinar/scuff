package db

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Ollinar/scuff/internal/search"

	"github.com/jmoiron/sqlx"
)

var (
	logger  = slog.Default()
	decimal = 10
)

func rollbackTxx(tx *sqlx.Tx) {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}
	if tx == nil {
		return
	}
	err := tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		logger.Error(err.Error())
	}
}

// TODO: refactor this so its only for used for read queries. write queries should explecitly use sqlx.tx
type dbtx interface {
	Get(dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	Rebind(q string) string
}

func parseIdsStr(s string) ([]int64, error) {
	ids := make([]int64, 0)
	if s != "" {
		for v := range strings.SplitSeq(s, ",") {
			id, err := strconv.ParseInt(v, decimal, 64)
			if err != nil {
				return nil, fmt.Errorf("unexpected error while parsing id string: %s , original error: %w", s, err)
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func DiffIdsSeq(newIds, oldIds []int, includeRemoved bool) iter.Seq2[int, int] {
	return diffIdsSeq(newIds, oldIds, includeRemoved)
}

// diffIdsSeq will return sequence of changed values from newIds to oldIds, the values are pair of index and the id.
// -1 index means it's been removed from oldIds
func diffIdsSeq[T cmp.Ordered](newIds []T, oldIds []T, includeRemoved bool) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range newIds {
			if i >= len(oldIds) || oldIds[i] != v {
				if !yield(i, v) {
					return
				}
			}
		}
		if !includeRemoved {
			return
		}
		// loop to look for those thas been removed
		for i, v := range oldIds {
			if i >= len(newIds) || newIds[i] != v {

				idx := slices.Index(newIds, v)
				if idx >= 0 {
					continue
				}
				if !yield(-1, v) {
					return
				}
			}
		}
	}
}

func extractIds[T any](s []T, fn func(T) int) []int {
	ids := make([]int, 0, len(s))
	for _, v := range s {
		ids = append(ids, fn(v))
	}
	return ids
}

func buildRegexpStringFilter(f search.StringFilter) string {
	v := regexp.QuoteMeta(f.Value)
	switch f.Type {
	case search.MatchingExact:
		return "(?i)^" + v + "$"
	case search.MatchingInfix:
		return "(?i)" + v
	case search.MatchingPrefix:
		return "(?i)^" + v
	case search.MatchingSuffix:
		return "(?i)" + v + "$"
	default:
		return "(?i)" + v
	}
}
