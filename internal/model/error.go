package model

import (
	"errors"
)

type appError error

var (
	ErrInvalidEntity appError = errors.New("entity is invalid")
	ErrUnexpected    appError = errors.New("unexpected error encountered")
	ErrNotFound      appError = errors.New("entity not found")
	ErrUnsupported   appError = errors.New("unsupported")
)
