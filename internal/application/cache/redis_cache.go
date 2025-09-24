package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

type RedisCache struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewRedisCache(address, password string, ttlSeconds int) interfaces.Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})
	return &RedisCache{
		redisClient: client,
		ttl:         time.Duration(ttlSeconds) * time.Second,
	}
}

func (c *RedisCache) GetUser(
	ctx context.Context,
	login string,
) (*entities.User, bool, error) {

	key := "user:" + login
	value, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var user entities.User
	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

func (c *RedisCache) SetUser(ctx context.Context, user *entities.User) error {
	key := "user:" + user.Login
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.redisClient.Set(ctx, key, bytes, c.ttl).Err()
}

func (c *RedisCache) DeleteUser(ctx context.Context, login string) error {
	key := "user:" + login
	return c.redisClient.Del(ctx, key).Err()
}
