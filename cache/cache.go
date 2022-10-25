package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	conn *redis.Client
	ttl  time.Duration
}

func Connect(ctx context.Context, addr string, ttl time.Duration) (*Cache, error) {
	opts := redis.Options{
		Addr: addr,
	}
	c := redis.NewClient(&opts)
	cache := Cache{c, ttl}

	if err := cache.Health(ctx); err != nil {
		return nil, err
	}

	return &cache, nil
}

func (c *Cache) Health(ctx context.Context) error {
	_, err := c.conn.Ping(ctx).Result()
	return err
}

var ErrNotFound = errors.New("not found")

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	v, err := c.conn.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return []byte(v), nil
}

func (c *Cache) Set(ctx context.Context, key string, value []byte) error {
	return c.conn.Set(ctx, key, value, c.ttl).Err()
}
