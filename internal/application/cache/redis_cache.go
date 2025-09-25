package cache

import (
	"context"
	"encoding/json"
	"log"
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
	log.Printf("[INFO] Redis client initialized at %s", address)
	return &RedisCache{
		redisClient: client,
		ttl:         time.Duration(ttlSeconds) * time.Second,
	}
}

func (cache *RedisCache) GetUser(ctx context.Context, login string) (*entities.User, bool, error) {
	key := "user:" + login
	value, err := cache.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("[INFO] Cache miss for key: %s", key)
		return nil, false, nil
	}
	if err != nil {
		log.Printf("[ERROR] Redis GET error for key %s: %v", key, err)
		return nil, false, err
	}

	var user entities.User
	if err := json.Unmarshal([]byte(value), &user); err != nil {
		log.Printf("[ERROR] Failed to unmarshal cache value for key %s: %v", key, err)
		return nil, false, err
	}

	log.Printf("[INFO] Cache hit for key: %s", key)
	return &user, true, nil
}

func (cache *RedisCache) SetUser(ctx context.Context, user *entities.User) error {
	key := "user:" + user.Login
	bytes, err := json.Marshal(user)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal user for cache key %s: %v", key, err)
		return err
	}

	if err := cache.redisClient.Set(ctx, key, bytes, cache.ttl).Err(); err != nil {
		log.Printf("[ERROR] Redis SET error for key %s: %v", key, err)
		return err
	}

	log.Printf("[INFO] User cached with key: %s", key)
	return nil
}

func (cache *RedisCache) DeleteUser(ctx context.Context, login string) error {
	key := "user:" + login
	if err := cache.redisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[ERROR] Redis DEL error for key %s: %v", key, err)
		return err
	}

	log.Printf("[INFO] Cache deleted for key: %s", key)
	return nil
}
