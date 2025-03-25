package cache

import "github.com/go-redis/redis/v8"

func NewRedisClient(addr, pw string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})
}

func NewRedisStorage(rdb *redis.Client, expiry *ExpiryTimes) Storage {
	return Storage{
		Users:          &UsersStore{db: rdb, expiry: expiry.Users},
		Contacts:       &ContactsStore{db: rdb, expiry: expiry.Contacts},
		Misc:           &MiscStore{db: rdb},
		EncryptionKeys: &EncryptionKeysStore{db: rdb, expiry: expiry.EncryptionKeys},
	}
}
