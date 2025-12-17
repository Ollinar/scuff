package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"

	"github.com/Ollinar/scuff/internal/repository"
)

func generatePartialHash(r io.Reader, hasher crypto.Hash, len int64) (string, error) {
	h := hasher.New()
	_, err := io.CopyN(h, r, len)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func encodeIdsToBinary(ids []int64) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(ids)*8))
	for _, v := range ids {
		err := binary.Write(buf, binary.NativeEndian, v)
		if err != nil {
			// should never happen but just incase
			panic("unexpected error encoding to binary: " + err.Error())
		}
	}
	return buf.Bytes()
}

func decodeIdsFromBinary(data []byte) []int64 {
	// size of int64 is 8
	idSize := 8
	dataLen := len(data) / idSize
	ids := make([]int64, 0, dataLen)
	buf := bytes.NewReader(data)
	var tmpInt int64
	for range dataLen {
		err := binary.Read(buf, binary.NativeEndian, &tmpInt)
		if err != nil {
			// should never happen but just incase
			panic("unexpected error encoding to binary: " + err.Error())
		}
		ids = append(ids, tmpInt)
	}
	return ids
}

func beginTransaction(ctx context.Context, repo any) (context.Context, error) {
	var err error
	if acd, ok := repo.(repository.Acid); ok {
		ctx, err = acd.WithTransaction(ctx)
	}

	return ctx, err
}

func saveTransaction(ctx context.Context, repo any) error {
	var err error
	if acd, ok := repo.(repository.Acid); ok {
		err = acd.Save(ctx)
		return err
	}
	return nil
}

func rollbackTransaction(ctx context.Context, repo any) error {
	var err error
	if acd, ok := repo.(repository.Acid); ok {
		err = acd.Rollback(ctx)
		return err
	}
	return nil
}

func reopenZipFile(file fs.File, zr *zip.ReadCloser, filename string) (fs.File, error) {
	err := file.Close()
	if err != nil {
		return nil, err
	}
	file, err = zr.Open(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func setFromSlice[T comparable](s []T) []T {
	tmpMp := make(map[T]struct{}, len(s))
	for _, v := range s {
		tmpMp[v] = struct{}{}
	}
	s = make([]T, 0, len(tmpMp))
	for el := range tmpMp {
		s = append(s, el)
	}
	return s
}
