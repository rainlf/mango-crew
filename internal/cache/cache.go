package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rainlf/mango-crew/internal/config"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Store 提供一个轻量缓存封装，便于在 service 层按 key 读写 JSON 数据。
type Store struct {
	client    *redis.Client
	enabled   bool
	keyPrefix string
}

func NewStore(cfg config.RedisConfig) (*Store, error) {
	store := &Store{
		enabled:   cfg.Enabled,
		keyPrefix: cfg.Prefix(),
	}
	if !cfg.Enabled {
		return store, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis failed: %w", err)
	}

	store.client = client
	return store, nil
}

func (s *Store) Enabled() bool {
	return s != nil && s.enabled && s.client != nil
}

func (s *Store) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	if !s.Enabled() {
		return false, nil
	}

	val, err := s.client.Get(ctx, s.buildKey(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		logger.Warn("cache unmarshal failed", logger.String("key", key), logger.Err(err))
		return false, err
	}
	return true, nil
}

func (s *Store) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.buildKey(key), data, ttl).Err()
}

func (s *Store) Delete(ctx context.Context, keys ...string) error {
	if !s.Enabled() || len(keys) == 0 {
		return nil
	}

	redisKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		redisKeys = append(redisKeys, s.buildKey(key))
	}
	return s.client.Del(ctx, redisKeys...).Err()
}

func (s *Store) DeleteByPrefix(ctx context.Context, prefixes ...string) error {
	if !s.Enabled() || len(prefixes) == 0 {
		return nil
	}

	for _, prefix := range prefixes {
		pattern := s.buildKey(prefix) + "*"
		var cursor uint64
		for {
			keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				if err := s.client.Del(ctx, keys...).Err(); err != nil {
					return err
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}
	return nil
}

func (s *Store) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *Store) buildKey(key string) string {
	return s.keyPrefix + ":" + key
}
