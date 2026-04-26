package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned by Get when the idempotency key does not exist in the store.
// Callers should distinguish this from a Redis connectivity error.
var ErrNotFound = errors.New("idempotency key not found")

type IdempotencyRecord struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

type IdempotencyStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewIdempotencyStore(client *redis.Client, ttl time.Duration) *IdempotencyStore {
	return &IdempotencyStore{client: client, ttl: ttl}
}

func (s *IdempotencyStore) Get(ctx context.Context, userID int64, key string) (*IdempotencyRecord, error) {
	data, err := s.client.Get(ctx, s.redisKey(userID, key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var rec IdempotencyRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (s *IdempotencyStore) Set(ctx context.Context, userID int64, key string, rec *IdempotencyRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.redisKey(userID, key), data, s.ttl).Err()
}

func (s *IdempotencyStore) redisKey(userID int64, key string) string {
	return fmt.Sprintf("idempotency:%d:%s", userID, key)
}
