# =========================
# Run (Local Development)
# =========================
.PHONY: run

run:
	go run cmd/server/main.go



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
	swag init \
		-g server/main.go \
		--dir ./cmd,./internal \
		--output docs
	go mod tidy
	GOOS=linux GOARCH=amd64 go build -o dealna ./cmd/server



# =========================
# Migrations
# =========================
MIGRATIONS_DIR := ../database/migrations

.PHONY: migrate-up migrate-down-1 migrate-down migrate-status migrate-create

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

migrate-create:
ifndef name
	$(error ❌ please provide a name for the migration. Example: make migrate-create name=create_users_table)
endif
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -format "20060102150405" $(name)

# =========================
# Swagger
# =========================
.PHONY: swagger swagger-clean

swagger:
	swag init \
		-g cmd/server/main.go \
		--dir . \
		--output docs

swagger-clean:
	rm -rf docs
