package data

import (
	"errors"
	"sort"
	"strconv"
)

type MemoryDB struct {
	m       map[string]string
	ordered map[string]list
}

type member struct {
	value string
	score int
}

type list []member

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{m: make(map[string]string), ordered: make(map[string]list)}
}

func (*MemoryDB) Ping() bool {
	return true
}

func (db *MemoryDB) Set(u KeyValue) error {
	db.m[u.Key()] = u.String()
	return nil
}

func (db *MemoryDB) Increment(key string) (int64, error) {
	i, err := strconv.ParseInt(db.m[key], 10, 0)
	if err != nil {
		return 0, err
	}

	result := i + 1
	db.m[key] = strconv.FormatInt(result, 10)
	return result, nil
}

func (db *MemoryDB) Get(key string) (string, error) {
	if val, ok := db.m[key]; ok {
		return val, nil
	}
	return "", errors.New("key does not exist")
}

func (db *MemoryDB) Remove(u string) error {
	delete(db.m, u)
	return nil
}

func (db *MemoryDB) SetOrdered(kv KeyValue, score int) error {
	m := member{kv.String(), score}
	if val, ok := db.ordered[kv.Key()]; ok {
		val = append(val, m)
		db.ordered[kv.Key()] = val
	} else {
		db.ordered[kv.Key()] = []member{m}
	}
	return nil
}

func (db *MemoryDB) GetAllOrderedByScore(key string) []string {
	if val, ok := db.ordered[key]; ok {
		sort.Slice(val, func(i, j int) bool {
			return val[i].score < val[j].score
		})
		return val.values()
	}
	return nil
}

func (db *MemoryDB) RemoveOrdered(kv KeyValue) error {
	delete(db.ordered, kv.Key())
	return nil
}

func (l list) values() []string {
	strings := make([]string, 0)
	for _, v := range l {
		strings = append(strings, v.value)
	}
	return strings
}
