package main

import (
	"github.com/9thDuck/chat_go.git/internal/auth"
	"github.com/9thDuck/chat_go.git/internal/store/cache"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	expiry      *cache.ExpiryTimes
}

type s3Cfg struct {
	cfg        *aws.Config
	bucketName string
}

type cloudCfg struct {
	s3 s3Cfg
}
type config struct {
	appName  string
	addr     string
	dbConfig dbConfig
	env      string
	cloud    cloudCfg
	auth     authConfig
	cacheCfg cacheCfg
}
