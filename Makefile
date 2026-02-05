# =========================
# Build (Local & CI SAFE)
# =========================
.PHONY: build

build:
	sqlc generate
	go mod tidy
	go build -o dealna ./cmd/server


# =========================
# Build for Linux (EC2 / CI)
# =========================
.PHONY: build-linux

build-linux:
	sqlc generate
	go mod tidy
	GOOS=linux GOARCH=amd64 go build -o dealna ./cmd/server


# =========================
# Config
# =========================
MIGRATIONS_DIR := migrations

ifndef DATABASE_URL
$(error ❌ DATABASE_URL is not set)
endif


# =========================
# Migration commands
# =========================
.PHONY: migrate-up migrate-down migrate-down-1 migrate-status migrate-create

## Apply all up migrations
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

## Roll back ONE migration only (SAFE)
migrate-down-1:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

## Roll back ALL migrations (DANGEROUS – intentional)
migrate-down:
	@echo "⚠️  This will rollback ALL migrations!"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

## Show current migration version
migrate-status:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

## Create a new migration
## usage: make migrate-create name=add_sessions_table
migrate-create:
ifndef name
	$(error ❌ name is required. Usage: make migrate-create name=add_sessions_table)
endif
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# =========================
# Swagger
# =========================
.PHONY: swagger swagger-clean

## Generate swagger docs
swagger:
	swag init -g cmd/server/main.go --output docs

## Remove generated swagger docs
swagger-clean:
	rm -rf docs
