package data

import (
	"io"
)

type MediaRepo interface {
	Store(io.Reader, string, string, int64) (string, error)
}
