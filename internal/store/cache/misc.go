package cache

import "github.com/go-redis/redis/v8"

type MiscStore struct {
	db *redis.Client
}

func (s *MiscStore) Close() error {
	return s.db.Close()
}
