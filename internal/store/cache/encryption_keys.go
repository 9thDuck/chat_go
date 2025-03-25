package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/go-redis/redis/v8"
)

type EncryptionKeysStore struct {
	db     *redis.Client
	expiry time.Duration
}

func (s *EncryptionKeysStore) Get(ctx context.Context, userID int64, encryptionKeyID string) (*store.EncryptionKey, error) {
	cacheKey := fmt.Sprintf("encryption_key:%d:%s", userID, encryptionKeyID)

	data, err := s.db.Get(ctx, cacheKey).Result()

	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var encryptionKey store.EncryptionKey
	if data == "" {
		return nil, store.ErrNotFound
	}

	return &encryptionKey, nil
}

func (s *EncryptionKeysStore) Set(ctx context.Context, userID int64, encryptionKey *store.EncryptionKey) error {
	cacheKey := fmt.Sprintf("encryption_key:%d:%s", userID, encryptionKey.ID)
	return s.db.Set(ctx, cacheKey, encryptionKey.Key, s.expiry).Err()
}

func (s *EncryptionKeysStore) Delete(ctx context.Context, userID int64, encryptionKeyID string) error {
	cacheKey := fmt.Sprintf("encryption_key:%d:%s", userID, encryptionKeyID)
	return s.db.Del(ctx, cacheKey).Err()
}
