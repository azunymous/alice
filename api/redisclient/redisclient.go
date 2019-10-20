package redisclient

import (
	"errors"
	"github.com/alice-ws/alice/data"
	"github.com/go-redis/redis"
)

type RedisClient struct {
	client *redis.Client
}

func (r *RedisClient) SetOrdered(kv data.KeyValue, score int) error {
	result := r.client.ZAdd(kv.Key(), redis.Z{
		Score:  float64(score),
		Member: kv.String(),
	})

	return result.Err()
}

func (r *RedisClient) GetAllOrderedByScore(key string) []string {
	slice := r.client.ZRevRangeByScore(key, redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	})
	strings, err := slice.Result()
	if err != nil {
		return nil
	}
	return strings
}

func (r *RedisClient) RemoveOrdered(kv data.KeyValue) error {
	result := r.client.ZRem(kv.Key(), kv.String())
	return result.Err()
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

func (r *RedisClient) Ping() bool {
	if err := r.client.Ping().Err(); err != nil {
		return false
	}
	return true
}

func (r *RedisClient) Set(kv data.KeyValue) error {
	_, err := r.client.Set(kv.Key(), kv.String(), 0).Result()
	return err
}

func (r *RedisClient) Increment(key string) (int64, error) {
	incr := r.client.Incr(key)
	return incr.Val(), incr.Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	result, getErr := r.client.Get(key).Result()
	if getErr != nil {
		return "", errors.New("error getting user " + getErr.Error())
	}
	return result, nil
}

func (r *RedisClient) Remove(key string) error {
	err := r.client.Del(key).Err()
	return err
}
