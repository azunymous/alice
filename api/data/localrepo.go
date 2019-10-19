package data

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type LocalRepo struct {
	dir string
}

func NewLocalRepo(dir string) MediaRepo {
	log.Println("Creating directory " + dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory %s, continuing", dir)
	}
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

func (r LocalRepo) GenerateUniqueName(fileName string) string {
	ext := path.Ext(fileName)
	return strconv.FormatInt(time.Now().UnixNano(), 10) + ext
}
