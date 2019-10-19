package data

import (
	"io"
	"io/ioutil"
	"path/filepath"
)

type LocalRepo struct {
	dir string
}

func NewLocalRepo(dir string) MediaRepo {
	return LocalRepo{dir: dir}
}

func (r LocalRepo) Store(fileReader io.Reader, _ string, name string, _ int64) (string, error) {

	tempImage, err := ioutil.TempFile(r.dir, "*-"+name)
	inMemoryImage, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return "", err
	}

	_, err = tempImage.Write(inMemoryImage)
	if err != nil {
		return "", err
	}

	return filepath.Base(tempImage.Name()), nil
}
