# =========================
# Build
# =========================
.PHONY: build build-linux

build:
	sqlc generate
	go mod tidy
	go build -o dealna ./cmd/server

build-linux:
	sqlc generate
	go mod tidy
	GOOS=linux GOARCH=amd64 go build -o dealna ./cmd/server


# =========================
# Migrations
# =========================
MIGRATIONS_DIR := ../database/migrations

.PHONY: migrate-up migrate-down-1 migrate-down migrate-status

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

swagger:
	swag init -g cmd/server/main.go --dir cmd/server,internal --output docs

swagger-clean:
	rm -rf docs
