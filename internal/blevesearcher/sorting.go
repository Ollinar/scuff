package blevesearcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/Ollinar/scuff/internal/search"

	"github.com/blevesearch/bleve/v2/numeric"
	bleveSearch "github.com/blevesearch/bleve/v2/search"
)

// NOTE: Taken from official implementation of SortField with some modification (mainly on filterTermsByMode)

type SortNamespace struct {
	Field     string
	Desc      bool
	Type      bleveSearch.SortFieldType
	Mode      bleveSearch.SortFieldMode
	Missing   bleveSearch.SortFieldMissing
	values    [][]byte
	tmp       [][]byte
	Namespace search.StringFilter
}

// UpdateVisitor notifies this sort field that in this document
// this field has the specified term
func (s *SortNamespace) UpdateVisitor(field string, term []byte) {
	if field == s.Field {
		s.values = append(s.values, term)
	}
}

// Value returns the sort value of the DocumentMatch
// it also resets the state of this SortField for
// processing the next document
func (s *SortNamespace) Value(i *bleveSearch.DocumentMatch) string {
	iTerms := s.filterTermsByType(s.values)
	iTerm := s.filterTermsByMode(iTerms)
	s.values = s.values[:0]
	return iTerm
}

func (s *SortNamespace) DecodeValue(value string) string {
	switch s.Type {
	case bleveSearch.SortFieldAsNumber:
		i64, err := numeric.PrefixCoded(value).Int64()
		if err != nil {
			return value
		}
		return strconv.FormatFloat(numeric.Int64ToFloat64(i64), 'f', -1, 64)
	case bleveSearch.SortFieldAsDate:
		i64, err := numeric.PrefixCoded(value).Int64()
		if err != nil {
			return value
		}
		return time.Unix(0, i64).UTC().Format(time.RFC3339Nano)
	default:
		return value
	}
}

// Descending determines the order of the sort
func (s *SortNamespace) Descending() bool {
	return s.Desc
}

func (s *SortNamespace) filterTermsByMode(terms [][]byte) string {
	if len(terms) == 1 || (len(terms) > 1 && s.Mode == bleveSearch.SortFieldDefault) {
		return string(terms[0])
	} else if len(terms) > 1 {
		nmcpsaceregex := regexp.MustCompile(fmt.Sprintf("(?i)^%s:", buildRegexpStringFilter(s.Namespace)))
		matchedTerm := make([][]byte, 0, len(terms))
		for _, v := range terms {
			if nmcpsaceregex.Match(v) {
				matchedTerm = append(matchedTerm, v)
			}
		}
		slices.SortStableFunc(matchedTerm, func(a, b []byte) int { return bytes.Compare(a, b) })

		switch s.Mode {
		case bleveSearch.SortFieldMin:
			if len(matchedTerm) > 0 {
				return string(matchedTerm[0])
			}
		case bleveSearch.SortFieldMax:
			if len(matchedTerm) > 0 {
				return string(matchedTerm[len(matchedTerm)-1])
			}
		}
	}

	// handle missing terms
	if s.Missing == bleveSearch.SortFieldMissingLast {
		if s.Desc {
			return bleveSearch.LowTerm
		}
		return bleveSearch.HighTerm
	}
	if s.Desc {
		return bleveSearch.HighTerm
	}
	return bleveSearch.LowTerm
}

// filterTermsByType attempts to make one pass on the terms
// if we are in auto-mode AND all the terms look like prefix-coded numbers
// return only the terms which had shift of 0
// if we are in explicit number or date mode, return only valid
// prefix coded numbers with shift of 0
func (s *SortNamespace) filterTermsByType(terms [][]byte) [][]byte {
	stype := s.Type

	switch stype {
	case bleveSearch.SortFieldAuto:
		allTermsPrefixCoded := true
		termsWithShiftZero := s.tmp[:0]
		for _, term := range terms {
			valid, shift := numeric.ValidPrefixCodedTermBytes(term)
			if valid && shift == 0 {
				termsWithShiftZero = append(termsWithShiftZero, term)
			} else if !valid {
				allTermsPrefixCoded = false
			}
		}
		// reset the terms only when valid zero shift terms are found.
		if allTermsPrefixCoded && len(termsWithShiftZero) > 0 {
			terms = termsWithShiftZero
			s.tmp = termsWithShiftZero[:0]
		}
	case bleveSearch.SortFieldAsNumber, bleveSearch.SortFieldAsDate:
		termsWithShiftZero := s.tmp[:0]
		for _, term := range terms {
			valid, shift := numeric.ValidPrefixCodedTermBytes(term)
			if valid && shift == 0 {
				termsWithShiftZero = append(termsWithShiftZero, term)
			}
		}
		terms = termsWithShiftZero
		s.tmp = termsWithShiftZero[:0]
	}

	return terms
}

// RequiresDocID says this SearchSort does not require the DocID be loaded
func (s *SortNamespace) RequiresDocID() bool { return false }

// RequiresScoring says this SearchStore does not require scoring
func (s *SortNamespace) RequiresScoring() bool { return false }

// RequiresFields says this SearchStore requires the specified stored field
func (s *SortNamespace) RequiresFields() []string { return []string{s.Field} }

func (s *SortNamespace) MarshalJSON() ([]byte, error) {
	// see if simple format can be used
	if s.Missing == bleveSearch.SortFieldMissingLast &&
		s.Mode == bleveSearch.SortFieldDefault &&
		s.Type == bleveSearch.SortFieldAuto {
		if s.Desc {
			return json.Marshal("-" + s.Field)
		}
		return json.Marshal(s.Field)
	}
	sfm := map[string]any{
		"by":    "field",
		"field": s.Field,
	}
	if s.Desc {
		sfm["desc"] = true
	}
	if s.Missing > bleveSearch.SortFieldMissingLast {
		switch s.Missing {
		case bleveSearch.SortFieldMissingFirst:
			sfm["missing"] = "first"
		}
	}
	if s.Mode > bleveSearch.SortFieldDefault {
		switch s.Mode {
		case bleveSearch.SortFieldMin:
			sfm["mode"] = "min"
		case bleveSearch.SortFieldMax:
			sfm["mode"] = "max"
		}
	}
	if s.Type > bleveSearch.SortFieldAuto {
		switch s.Type {
		case bleveSearch.SortFieldAsString:
			sfm["type"] = "string"
		case bleveSearch.SortFieldAsNumber:
			sfm["type"] = "number"
		case bleveSearch.SortFieldAsDate:
			sfm["type"] = "date"
		}
	}

	return json.Marshal(sfm)
}

func (s *SortNamespace) Copy() bleveSearch.SearchSort {
	rv := *s
	return &rv
}

func (s *SortNamespace) Reverse() {
	s.Desc = !s.Desc
	if s.Missing == bleveSearch.SortFieldMissingFirst {
		s.Missing = bleveSearch.SortFieldMissingLast
	} else {
		s.Missing = bleveSearch.SortFieldMissingFirst
	}
}
