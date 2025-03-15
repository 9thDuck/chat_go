package cache

import "github.com/go-redis/redis/v8"

func NewRedisClient(addr, pw string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users:    &UsersStore{db: rdb},
		Contacts: &ContactsStore{db: rdb},
		Misc:     &MiscStore{db: rdb},
	}
}
