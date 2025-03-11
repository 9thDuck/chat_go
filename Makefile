include .env

MIGRATIONS_PATH=./cmd/migrate/migrations

.PHONY: migrate-create
migration:
	~/Applications/go-migrate/migrate create -seq -ext sql -dir ${MIGRATIONS_PATH} $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	~/Applications/go-migrate/migrate -path=${MIGRATIONS_PATH} -database=${DB_ADDR} up

.PHONY: migrate-down
migrate-down:
	~/Applications/go-migrate/migrate -path=${MIGRATIONS_PATH} -database=${DB_ADDR} down
