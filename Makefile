.PHONY: up down backend-test frontend-test dev-backend

up:
	docker compose up --build

down:
	docker compose down -v

# Run Go tests (requires Go installed)
backend-test:
	cd backend && go test ./...

# Run frontend tests (requires Node installed)
frontend-test:
	cd frontend && npm test

# Run backend locally against Docker Postgres
dev-backend:
	cd backend && \
	DB_HOST=localhost DB_PORT=5432 DB_USER=collabb DB_PASSWORD=collabb DB_NAME=collabb \
	JWT_SECRET=dev go run ./cmd/server
