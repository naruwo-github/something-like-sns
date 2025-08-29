-include .env
export

proto:
	cd packages/protos && buf generate

migrate:
	migrate -path packages/dbschema/migrations -database "mysql://$${DB_USER}:$${DB_PASS}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}" up

seed:
	cd apps/api && go run ./cmd/seed/main.go

api-dev:
	cd apps/api && go run ./cmd/server

web-dev:
	pnpm --filter @repo/web dev
