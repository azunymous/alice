package data

import (
	"errors"
	"sort"
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

func (s *MemoryDB) Set(u KeyValue) error {
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

func (o *MemoryDB) SetOrdered(kv KeyValue, score int) error {
	m := member{kv.String(), score}
	if val, ok := o.ordered[kv.Key()]; ok {
		val = append(val, m)
	} else {
		o.ordered[kv.Key()] = []member{m}
	}
	return nil
}

func (o *MemoryDB) GetAllOrderedByScore(key string) []string {
	if val, ok := o.ordered[key]; ok {
		sort.Slice(val, func(i, j int) bool {
			return val[i].score < val[j].score
		})
		return val.values()
	}
	return nil
}

func (o *MemoryDB) RemoveOrdered(kv KeyValue) error {
	delete(o.ordered, kv.Key())
	return nil
}

func (l list) values() []string {
	strings := make([]string, len(l))
	for _, v := range l {
		strings = append(strings, v.value)
	}
	return strings
}
