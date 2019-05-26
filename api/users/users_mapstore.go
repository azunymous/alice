package users

import (
	"errors"
)

type MemoryStore struct {
	m map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{m: make(map[string]string)}
}

func (*MemoryStore) Ping() bool {
	return true
}

func (s *MemoryStore) Add(u User) error {
	s.m[u.Username] = u.String()
	return nil
}

func (s *MemoryStore) Get(username string) (User, error) {
	if val, ok := s.m[username]; ok {
		user, e := NewUser(val)
		if e != nil {
			return User{}, e
		}
		return *user, nil
	}
	return User{}, errors.New("user does not exist")
}

func (s *MemoryStore) Remove(u User) error {
	delete(s.m, u.Username)
	return nil
}
