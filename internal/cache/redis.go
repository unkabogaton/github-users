package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/unkabogaton/github-users/internal/models"
)

type RedisCache struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewRedisCache(address, password string, ttlSeconds int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})
	return &RedisCache{
		redisClient: client,
		ttl:         time.Duration(ttlSeconds) * time.Second,
	}
}

func (cache *RedisCache) GetUser(
	ctx context.Context,
	login string,
) (*models.GitHubUser, bool, error) {

	key := "user:" + login
	value, err := cache.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var user models.GitHubUser
	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

func (cache *RedisCache) SetUser(ctx context.Context, user *models.GitHubUser) error {
	key := "user:" + user.Login
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return cache.redisClient.Set(ctx, key, bytes, cache.ttl).Err()
}

func (cache *RedisCache) DeleteUser(ctx context.Context, login string) error {
	key := "user:" + login
	return cache.redisClient.Del(ctx, key).Err()
}
