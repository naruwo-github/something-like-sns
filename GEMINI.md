## Project Overview

This is a boilerplate project for a multi-tenant SNS application, built with a Go backend and a Next.js frontend in a pnpm/Turborepo monorepo. It serves as a minimal, runnable example of a system with row-level database tenancy.

## Documentation Structure

This project has three main documentation files with distinct purposes:

1.  **`README.md`**: Provides a quick start guide for developers. It contains the essential steps to get the project running locally, an overview of the tech stack, and the basic repository structure. **Start here if you want to run the project.**

2.  **`SOFTWARE_DESIGN.md`**: This is the comprehensive design document. It contains all the detailed information about the project's architecture, data models (DDL), API specifications (.proto definitions), multi-tenancy strategy, and development conventions. **Refer to this for any in-depth questions about how the system is designed or why certain decisions were made.**

3.  **`AWS_ARCHITECTURE.md`**: This document outlines the system architecture for deploying the application to AWS, including infrastructure, CI/CD, and operational design.

## Key Technologies

*   **Backend**: Go, connect-go (RPC), Echo, `database/sql` (MySQL), golang-migrate
*   **Frontend**: Next.js, React, React Hooks (`useState`/`useEffect`)
*   **Monorepo**: pnpm, Turborepo
*   **API Definition**: Protocol Buffers (`.proto`)
*   **Dev Environment**: Docker Compose, Makefile

## Building and Running

All common tasks are defined as targets in the `Makefile`. For a step-by-step guide on how to get the development environment running, see the **"ローカル起動手順" (Local Setup)** section in the `README.md`.

## Key Files and Directories

*   `README.md`: Quick start guide.
*   `SOFTWARE_DESIGN.md`: Detailed design and architecture.
*   `AWS_ARCHITECTURE.md`: System architecture design for deploying to AWS.
*   `Makefile`: Defines common tasks like `proto`, `migrate`, `seed`, `api-dev`, `web-dev`.
*   `apps/api/`: The Go backend application.
*   `apps/web/`: The Next.js frontend application.
*   `packages/protos/`: The Protobuf files that define the API contracts.
*   `packages/dbschema/`: Contains the database schema and migration files.
*   `infra/local/docker-compose.yml`: Defines the local development services (MySQL, Adminer).