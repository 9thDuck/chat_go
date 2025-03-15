package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type ContactsStore struct {
	db     *redis.Client
	expiry time.Duration
}

func (s *ContactsStore) GetContactExists(ctx context.Context, userID, contactID int64) (bool, error) {
	minID, maxID := min(userID, contactID), max(userID, contactID)
	cacheKey := fmt.Sprintf("contact:%d:%d", minID, maxID)

	val, err := s.db.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return val == "1", nil
}

func (s *ContactsStore) SetContactExists(ctx context.Context, userID, contactID int64, exists bool) error {
	minID, maxID := min(userID, contactID), max(userID, contactID)
	cacheKey := fmt.Sprintf("contact:%d:%d", minID, maxID)

	val := "0"
	if exists {
		val = "1"
	}

	return s.db.SetEX(ctx, cacheKey, val, s.expiry).Err()
}

func (s *ContactsStore) DeleteContactExists(ctx context.Context, userID, contactID int64) error {
	minID, maxID := min(userID, contactID), max(userID, contactID)
	cacheKey := fmt.Sprintf("contact:%d:%d", minID, maxID)

	return s.db.Del(ctx, cacheKey).Err()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
