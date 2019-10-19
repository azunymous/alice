package data

import (
	"io"
)

type MediaRepo interface {
	Store(file io.Reader, group string, ID string, size int64) (URI string, err error)
	GenerateUniqueName(fileName string) string
}
