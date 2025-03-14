package main

import (
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
	token tokenConfig
}

type redisCfg struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type cacheCfg struct {
	initialised bool
	redis       redisCfg
}
type config struct {
	appName  string
	addr     string
	dbConfig dbConfig
	env      string
	auth     authConfig
	cacheCfg cacheCfg
}
