package redisclient

import (
	"errors"
	"github.com/alice-ws/alice/users"
	"github.com/go-redis/redis"
)

type Store struct {
	client *redis.Client
}

func ConnectToRedis(addr string) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &Store{client}, nil
}

func (s *Store) Ping() bool {
	if err := s.client.Ping().Err(); err != nil {
		return false
	}
	return true
}

func (s *Store) Add(u users.User) error {
	_, err := s.client.Set(u.Username, u.String(), 0).Result()
	return err
}

func (s *Store) Get(username string) (users.User, error) {
	result, getErr := s.client.Get(username).Result()
	if getErr != nil {
		return users.User{}, errors.New("error getting user " + getErr.Error())
	}
	user, newUserErr := users.NewUser(result)
	if newUserErr != nil {
		return users.User{}, errors.New("error reading stored user " + newUserErr.Error())
	}
	return *user, nil
}

func (s *Store) Remove(u users.User) error {
	err := s.client.Del(u.Username).Err()
	return err
}
