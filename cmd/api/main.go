package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/9thDuck/chat_go.git/internal/auth"
	"github.com/9thDuck/chat_go.git/internal/db"
	"github.com/9thDuck/chat_go.git/internal/env"
	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	godotenv.Load()
	DB_ADDR := os.Getenv("DB_ADDR")
	if DB_ADDR == "" {
		log.Fatal("DB_ADDR is missing in .env")
	}

	base64EncPrivKey := env.GetEnvString("JWT_EDDSA_PRIVATE_KEY", "")
	base64EncPubKey := env.GetEnvString("JWT_EDDSA_PUBLIC_KEY", "")

	if base64EncPrivKey == "" || base64EncPubKey == "" {
		log.Panic(`missing JWT_EDDSA_PRIVATE_KEY and/or JWT_EDDSA_PUBLIC_KEY in .env`)
	}

	privKeyByteArr, err := base64.StdEncoding.DecodeString(base64EncPrivKey)
	if err != nil {
		log.Panic("Malformed private key given in .env")
	}
	publicKeyByteArr, err := base64.StdEncoding.DecodeString(base64EncPubKey)
	if err != nil {
		log.Panic("Malformed public key given in .env")

	}

	conf :=
		config{
			addr: fmt.Sprintf(":%d", env.GetEnvInt("PORT", 8080)),
			dbConfig: dbConfig{addr: DB_ADDR,
				maxOpenConnections: env.GetEnvInt("DB_MAX_OPEN_CONNS", 30),
				maxIdleConnections: env.GetEnvInt("DB_MAX_IDLE_CONNS", 30),
				maxIdleTime:        env.GetEnvString("DB_MAX_IDLE_TIME", "15m"),
			},
			auth: authConfig{
				token: tokenConfig{
					keys: auth.EddsaKeys{
						Private: privKeyByteArr,
						Public:  publicKeyByteArr,
					},
					exp: auth.ExpiryDurations{
						Access:  time.Duration(env.GetEnvInt("JWT_ACCESS_TOKEN_EXPIRY_IN_MINS", 5)) * time.Minute,
						Refresh: time.Duration(env.GetEnvInt("JWT_REFRESH_TOKEN_EXPIRY_IN_DAYS", 7)) * time.Hour * 24,
					}},
			},
			appName: env.GetEnvString("APP_NAME", "DuckChat"),
		}

	jwtAuthenticator :=
		auth.NewJWTAutheticatorWithEddsa(
			conf.auth.token.keys,
			conf.appName,
			conf.appName,
		)

	dbConn, err := db.New(
		conf.dbConfig.addr,
		conf.dbConfig.maxOpenConnections,
		conf.dbConfig.maxIdleConnections,
		conf.dbConfig.maxIdleTime,
	)

	if err != nil {
		log.Fatal("db connection err", err)
	}

	defer dbConn.Close()

	store := store.NewStorage(dbConn)

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	app := &application{
		config:        conf,
		store:         store,
		logger:        logger,
		authenticator: jwtAuthenticator,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
