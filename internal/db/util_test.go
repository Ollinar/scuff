package db_test

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/Ollinar/scuff/internal/db"

	"github.com/jmoiron/sqlx"
)

type dbOpts func(*sqlx.DB)

func getDNS(enforceFK bool) string {
	dns := "./../../test/test.db"
	if enforceFK {
		dns += "?_fk=true"
	}
	return dns
}

func shuffleString(s string) string {
	r := []rune(s)
	rand.Shuffle(len(r), func(i, j int) {
		r[i], r[j] = r[j], r[i]
	})
	return string(r)
}

func shuffleElements[T any](s []T) []T {
	rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
	return s
}

func addNewRandomInt(s []int64) []int64 {
	for {
		newNum := rand.Int64()
		if !slices.Contains(s, newNum) {
			s = append(s, newNum)
			break
		}
	}
	return s
}

func removeRandomElement[T any](s []T) []T {
	inx := rand.IntN(len(s))
	tmps := make([]T, 0, len(s))
	for i, v := range s {
		if i == inx {
			continue
		}
		tmps = append(tmps, v)
	}
	return tmps
}

func TestDiffIdsSeq(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		newIds []int
		oldIds []int
		want   [][2]int
	}{
		// TODO: Add test cases.
		{
			name: "normal",
			oldIds: []int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
			},
			newIds: []int{
				1, 2, 3, 5, 7, 4, 9, 10,
			},
			want: [][2]int{
				{3, 5},
				{4, 7},
				{5, 4},
				{6, 9},
				{7, 10},
				{-1, 6},
				{-1, 8},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.DiffIdsSeq(tt.newIds, tt.oldIds, true)
			// TODO: update the condition below to compare got with tt.want.

			var res [][2]int
			for i, v := range got {
				res = append(res, [2]int{i, v})
			}

			eq := slices.EqualFunc(res, tt.want, func(e1, e2 [2]int) bool {
				return e1[0] == e2[0] && e1[1] == e2[1]
			})

			t.Logf("DiffIdsSeq() = %v, want %v", res, tt.want)
			if !eq {
				t.Errorf("DiffIdsSeq() = %v, want %v", got, tt.want)
			}
		})
	}
}
