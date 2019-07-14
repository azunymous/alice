package data

import (
	"errors"
)

type MemoryDB struct {
	m map[string]string
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{m: make(map[string]string)}
}

func (*MemoryDB) Ping() bool {
	return true
}

func (s *MemoryDB) Add(u KeyValue) error {
	s.m[u.Key()] = u.String()
	return nil
}

func (s *MemoryDB) Get(key string) (string, error) {
	if val, ok := s.m[key]; ok {
		return val, nil
	}
	return "", errors.New("key does not exist")
}

func (s *MemoryDB) Remove(u string) error {
	delete(s.m, u)
	return nil
}
