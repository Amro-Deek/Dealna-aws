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

# =========================
# Migration commands
# =========================
.PHONY: migrate-up migrate-down migrate-down-1 migrate-status migrate-create

migrate-up:
ifndef DATABASE_URL
	$(error ❌ DATABASE_URL is not set)
endif
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down-1:
ifndef DATABASE_URL
	$(error ❌ DATABASE_URL is not set)
endif
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-down:
ifndef DATABASE_URL
	$(error ❌ DATABASE_URL is not set)
endif
	@echo "⚠️  This will rollback ALL migrations!"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

migrate-status:
ifndef DATABASE_URL
	$(error ❌ DATABASE_URL is not set)
endif
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version


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
