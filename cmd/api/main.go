package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/9thDuck/chat_go.git/internal/db"
	"github.com/9thDuck/chat_go.git/internal/env"
	"github.com/9thDuck/chat_go.git/internal/logger"
	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	slog.SetDefault(logger.Logger)
	DB_ADDR := os.Getenv("DB_ADDR")
	if DB_ADDR == "" {
		log.Fatal("DB_ADDR is missing in .env")
	}

	conf :=
		config{
			addr: fmt.Sprintf(":%d", env.GetEnvInt("PORT", 8080)),
			dbConfig: dbConfig{addr: DB_ADDR,
				maxOpenConnections: env.GetEnvInt("DB_MAX_OPEN_CONNS", 30),
				maxIdleConnections: env.GetEnvInt("DB_MAX_IDLE_CONNS", 30),
				maxIdleTime:        env.GetEnvString("DB_MAX_IDLE_TIME", "15m"),
			},
		}

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
	app := &application{config: conf, store: store}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
