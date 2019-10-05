package redisclient

import (
	"errors"
	"github.com/alice-ws/alice/data"
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

func (s *Store) Set(kv data.KeyValue) error {
	_, err := s.client.Set(kv.Key(), kv.String(), 0).Result()
	return err
}

func (s *Store) Get(key string) (string, error) {
	result, getErr := s.client.Get(key).Result()
	if getErr != nil {
		return "", errors.New("error getting user " + getErr.Error())
	}
	return result, nil
}

func (s *Store) Remove(key string) error {
	err := s.client.Del(key).Err()
	return err
}
