package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const oidcStateKeyPrefix = "treepage:oidc:state:"

type redisStateStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStateStore(addr, password string, db int) (*redisStateStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &redisStateStore{client: client, ttl: 10 * time.Minute}, nil
}

func (s *redisStateStore) Create() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.client.Set(ctx, oidcStateKeyPrefix+state, "1", s.ttl).Err(); err != nil {
		return "", err
	}
	return state, nil
}

func (s *redisStateStore) Validate(state string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := oidcStateKeyPrefix + state
	n, err := s.client.Del(ctx, key).Result()
	return err == nil && n > 0
}

func NewStateStoreFromEnv(fallback StateStore) StateStore {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return fallback
	}
	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			db = n
		}
	}
	store, err := NewRedisStateStore(addr, os.Getenv("REDIS_PASSWORD"), db)
	if err != nil {
		return fallback
	}
	return store
}
