package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type TokenVersionRepository interface {
	Incr(ctx context.Context, userID uint64) (int, error)
	Get(ctx context.Context, userID uint64) (int, error)
}

type tokenVersionRepository struct {
	client *redis.Client
}

func NewTokenVersionRepository(client *redis.Client) TokenVersionRepository {
	return &tokenVersionRepository{client: client}
}

func key(userID uint64) string {
	return fmt.Sprintf("user:token_version:%d", userID)
}

func (r *tokenVersionRepository) Incr(ctx context.Context, userID uint64) (int, error) {
	result, err := r.client.Incr(ctx, key(userID)).Result()
	if err != nil {
		return 0, err
	}
	return int(result), nil
}

func (r *tokenVersionRepository) Get(ctx context.Context, userID uint64) (int, error) {
	val, err := r.client.Get(ctx, key(userID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return val, nil
}
