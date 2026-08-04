MIGRATIONS_PATH="db/migration"
DATABASE_URL=postgres://postgres:postgres@localhost:5432/new?sslmode=disable

###################
# Database        #
###################
.PHONY: mig-up
mig-up: ## Runs the migrations up
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

.PHONY: mig-down
mig-down: ## Runs the migrations down
	migrate -path ${MIGRATIONS_PATH} -database "$(DATABASE_URL)" down

.PHONY: new-mig
new-mig:
	migrate create -ext sql -dir ${MIGRATIONS_PATH} -seq $(NAME)

###################
# Linting         #
###################
.PHONY: lint-fix
lint-fix:
	go fmt ./...
	golangci-lint run --fix ./...

.PHONY: lint
lint:
	go fmt ./...
	golangci-lint run ./...

###################
# Testing         #
###################
.PHONY: mock
mock:
	mockery --output test/billing/mocks --dir internal/billing --all && \
	mockery --output test/shared/mocks --dir internal/shared --all

.PHONY: test
test: mock
	go test ./... -race
