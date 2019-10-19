package redisclient

import (
	"errors"
	"github.com/alice-ws/alice/data"
	"github.com/go-redis/redis"
)

type RedisClient struct {
	client *redis.Client
}

func ConnectToRedis(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisClient{client}, nil
}

func (s *RedisClient) Ping() bool {
	if err := s.client.Ping().Err(); err != nil {
		return false
	}
	return true
}

func (s *RedisClient) Set(kv data.KeyValue) error {
	_, err := s.client.Set(kv.Key(), kv.String(), 0).Result()
	return err
}

func (s *RedisClient) Get(key string) (string, error) {
	result, getErr := s.client.Get(key).Result()
	if getErr != nil {
		return "", errors.New("error getting user " + getErr.Error())
	}
	return result, nil
}

func (s *RedisClient) Remove(key string) error {
	err := s.client.Del(key).Err()
	return err
}
