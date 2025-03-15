package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/9thDuck/chat_go.git/internal/store"

	"github.com/go-redis/redis/v8"
)

type UsersStore struct {
	db     *redis.Client
	expiry time.Duration
}

func (s *UsersStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%d", userID)

	data, err := s.db.Get(ctx, cacheKey).Result()

	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	user.Role = &store.Role{}
	if data == "" {
		return nil, store.ErrNotFound
	}

	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UsersStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%d", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.db.SetEX(ctx, cacheKey, json, s.expiry).Err()
}

func (s *UsersStore) Delete(ctx context.Context, userID int64) error {
	cacheKey := fmt.Sprintf("user-%d", userID)
	return s.db.Del(ctx, cacheKey).Err()
}
