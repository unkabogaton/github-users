package cache

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	"github.com/unkabogaton/github-users/internal/domain/entities"
)

func TestRedisCache_SetGetDeleteUser(t *testing.T) {
	t.Parallel()
	mini, err := miniredis.Run()
	require.NoError(t, err)
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	cache := &RedisCache{redisClient: client, ttl: 0}

	ctx := context.Background()
	user := &entities.User{ID: 1, Login: "sample_username"}

	require.NoError(t, cache.SetUser(ctx, user))

	got, hit, err := cache.GetUser(ctx, "sample_username")
	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, user.Login, got.Login)

	require.NoError(t, cache.DeleteUser(ctx, "sample_username"))
	got, hit, err = cache.GetUser(ctx, "sample_username")
	require.NoError(t, err)
	require.False(t, hit)
	require.Nil(t, got)
}
