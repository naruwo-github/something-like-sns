-include .env
export
API_PORT ?= 8080

proto:
	cd packages/protos && buf generate

migrate:
	migrate -path packages/dbschema/migrations -database "mysql://$${DB_USER}:$${DB_PASS}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}" up

seed:
	cd apps/api && go run ./cmd/seed/main.go

api-dev:
	cd apps/api && (GO111MODULE=on go run github.com/air-verse/air@v1.52.2 || go run ./cmd/server)

api-kill:
	lsof -tiTCP:$(API_PORT) | xargs -r kill -9 || true

web-dev:
	pnpm --filter @repo/web dev
