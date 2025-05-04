# ===== General =====

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/bms ./cmd/server/main.go

# ===== Dev (auto-reload) =====

ENV_FILE := .env
ifneq ("$(wildcard $(ENV_FILE))","")
  include $(ENV_FILE)
  export
endif

PORT ?= 3000

dev:
	@echo "🔪 Killing port $(PORT) (if any)..."
	@lsof -ti :$(PORT) | xargs kill -9 || true
	@echo "🚀 Starting dev server with air..."
	@sh -c 'air -c .air.toml'


# ===== Atlas DB Migrations =====

inspect:
	go run -mod=mod ariga.io/atlas-provider-gorm load \
		--path ./internal/models \
		--dialect postgres > schema.hcl

diff:
	atlas migrate diff "change_$(shell date +%s)" --env gorm

apply:
	atlas migrate apply --url "postgres://bms:LyWZJj9mREOi2qCdq7ni2T17BS2@localhost:5432/bms?sslmode=disable"

clean-db:
	atlas schema clean --url "postgres://bms:LyWZJj9mREOi2qCdq7ni2T17BS2@localhost:5432/bms?sslmode=disable"

# ===== Format & Tidy =====

fmt:
	go fmt ./...

tidy:
	go mod tidy
