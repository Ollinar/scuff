package api

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/Ollinar/scuff/internal/model"
)

// generatePadedSubDir will generate the sub dirs and return the path to it, along side the length of padding used.
func generatePadedSubDir(pageID model.ID) (int, string) {
	idPerDir := int64(1000)
	start := (int64(pageID) / idPerDir) * idPerDir
	end := start + idPerDir - 1
	padding := len(fmt.Sprintf("%d", end))
	return padding, fmt.Sprintf("%0*d-%0*d", padding, start, padding, end)
}

type cachedPage struct {
	Data    []byte
	Modtime time.Time
}

func parseCachedPage(data []byte) (cachedPage, error) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	var parseedData cachedPage
	err := dec.Decode(&parseedData)
	if err != nil {
		return cachedPage{}, err
	}
	return parseedData, nil
}

func encodeCachedPage(pageData []byte, modtime time.Time) ([]byte, error) {
	data := cachedPage{
		Data:    pageData,
		Modtime: modtime,
	}
	buf := bytes.NewBuffer(make([]byte, 0, len(pageData)))
	err := gob.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
