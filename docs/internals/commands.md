# Commands

- Local Infrastructure: `docker compose up -d` (matches `.env.example`).
- Dev server: `air` (builds `./cmd/api/main.go`, hot reload).
- Tests: `task test` — `go test ./... -race`. Integration tests spin postgres containers, so docker must be available (or set `TEST_DATABASE_URL`).
- Lint: `task lint` (strict config in `.golangci.yml`; `task lint-fix` to auto-fix).
- Migrations: `task new-mig NAME=<name>` / `task mig-up` / `task mig-down` on `db/migration`.
- All other scripts live in the `Taskfile.yml`.
