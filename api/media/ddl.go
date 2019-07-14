package media

import (
	"encoding/json"
	"errors"
	"github.com/alice-ws/alice/data"
	"net/url"
)

type Store struct {
	db data.DB
}

func NewStore(db data.DB) *Store {
	return &Store{db: db}
}

type DDL struct {
	ID   string            `json:"id"`
	URL  url.URL           `json:"url"`
	Auth map[string]string `json:"auth"`
}

func (d *DDL) Key() string {
	return d.ID
}

func (d *DDL) String() (string, error) {
	bytes, err := json.Marshal(d)
	if err != nil {
		return "", errors.New("cannot convert DDL to json")
	}
	return string(bytes), nil
}

func newDDL(mjson string) (DDL, error) {
	d := DDL{}
	err := json.Unmarshal([]byte(mjson), &d)
	if err != nil {
		return DDL{}, errors.New("cannot parse json" + err.Error())
	}
	return d, nil
}
