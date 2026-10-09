# Commands

Source of truth is `Taskfile.yml` — this file summarizes; do not treat it as inventory.

- Local Infrastructure: `docker compose up -d` (matches `.env.example`). Shortcut: `task compose`.
- Dev server: `task` (runs `go mod tidy` + `air`, which builds `./cmd/api/main.go` with hot reload).
- Tests: `task test` — `go test ./... -race`. Integration tests need a real golden source (container, or `TEST_DATABASE_URL`).
- Lint: `task lint` (strict config in `.golangci.yml`; `task lint-fix` to auto-fix).
- Migrations: `task new-mig NAME=<name>` / `task mig-up` / `task mig-down` on `db/migration`.
- Build: `task build` — compiles the API binary to `bin/api`.
