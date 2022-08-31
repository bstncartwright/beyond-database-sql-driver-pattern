package cache

import (
	"context"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type Redis struct {
	client *redis.Client
}

func NewRedis() Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &Redis{client}
}

func (c *Redis) Get(key string) (any, bool) {
	v, err := c.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, false
	}

	return v, true
}

func (c *Redis) Put(key string, value any) error {
	return c.client.Set(context.Background(), key, value, time.Hour*300).Err()
}
