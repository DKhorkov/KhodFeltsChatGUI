package errors

import "errors"

var (
	ErrFileNotFound       = errors.New("file not found")
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrFileTooLarge       = errors.New("file too large")
)
