package main

import (
	"time"

	"github.com/9thDuck/chat_go.git/internal/auth"
)

type dbConfig struct {
	addr               string
	maxOpenConnections int
	maxIdleConnections int
	maxIdleTime        string
}

type tokenConfig struct {
	keys auth.EddsaKeys
	exp  auth.ExpiryDurations
}

type authConfig struct {
	basic basicAuthConfig
	token tokenConfig
}
