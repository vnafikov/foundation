package uploadhelper

import "errors"

var (
	ErrReadFileHeader = errors.New("cannot read file header")

	ErrEmptyFile = errors.New("empty file")

	ErrUnsupportedMIMEType = errors.New("unsupported MIME type")
)
