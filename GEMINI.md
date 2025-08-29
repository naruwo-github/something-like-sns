## Project Overview

This repository contains a boilerplate for a multi-tenant SNS (Social Networking Service) application. It is a monorepo managed with pnpm and Turborepo, containing a Go backend, a Next.js frontend, and shared Protobuf definitions.

The application supports features like tenants, users, posts, comments, reactions, and direct messages. It uses a shared database with row-level multi-tenancy.

**Key Technologies:**

*   **Backend:** Go, connect-go (for RPC), Echo (for web framework), Bun ORM, MySQL
*   **Frontend:** Next.js, React, TanStack Query, Tailwind CSS
*   **Monorepo:** pnpm, Turborepo
*   **API:** Protocol Buffers and connect-go for RPC.
*   **Database:** MySQL, with migrations handled by golang-migrate.
*   **Development:** Docker Compose, Makefile

## Building and Running

The project uses a `Makefile` to simplify common tasks.

**1. Initial Setup:**

Install dependencies:

```bash
pnpm install
brew install buf buildifier || true
```

**2. Generate Protobuf Code:**

Generate Go and TypeScript code from the `.proto` files:

```bash
make proto
```

**3. Start Database and Run Migrations:**

Start the MySQL database using Docker Compose, then run database migrations and seed the database:

```bash
docker compose -f infra/local/docker-compose.yml up -d
make migrate
make seed
```

**4. Run Development Servers:**

Run the API and web servers in separate terminals:

```bash
# Start the Go API server (on port 8080)
make api-dev

# Start the Next.js web server (on port 3000)
make web-dev
```

**5. Accessing the Application:**

*   **Web:** `http://acme.localhost:3000/`
*   **Adminer:** `http://localhost:8081/`

## Development Conventions

*   **Monorepo:** The project is a monorepo using pnpm workspaces and Turborepo.
*   **API:** The API is defined using Protocol Buffers in the `packages/protos` directory. Code generation is handled by `buf`.
*   **Database:** Database schema and migrations are located in `packages/dbschema`.
*   **Multi-tenancy:** The application uses a shared database with row-level multi-tenancy. The `tenant_id` is used to scope all data access.
*   **Testing:** The project includes unit, API integration, and E2E tests. The `README.md` mentions Playwright for E2E testing.
*   **Linting and Formatting:** The project uses ESLint and Prettier for the frontend, and `gofmt`/`golangci-lint` for the backend. You can run linting with `pnpm -w lint`.
