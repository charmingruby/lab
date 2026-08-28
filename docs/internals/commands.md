# Commands

- Local Infrastructure: `docker compose up -d` (matches `.env.example`).
- Dev server: `air` (builds `./cmd/api/main.go`, hot reload).
- Tests: `task test` — regenerates mocks first, so it needs `mockery` installed; then `go test ./... -race`.
- Lint: `task lint` (strict config in `.golangci.yml`; `task lint-fix` to auto-fix).
- Migrations: `task new-mig NAME=<name>` / `task mig-up` / `task mig-down` on `db/migration`.
- All other scripts live in the `Taskfile.yml`.
