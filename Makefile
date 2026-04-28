up:
	docker compose -f infra/compose.yml up -d

down:
	docker compose -f infra/compose.yml down

test-unit:
	go test -v tests/*_test.go

integration-test:
	ENV=test go test -v tests/integration/*_test.go
migrate:
	migrate -path infra/migrations -database "postgres://postgres:secret@localhost:5432/auth?sslmode=disable" -verbose up

cleandb:
	docker exec -i postgres-database psql -U postgres -d auth -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

dev:
	go run cmd/server/main.go