set dotenv-filename := ".env.local"

default:
  just --choose

test:
  RUN_INTEGRATION_TESTS=true go test ./...

integration-test-deps:
  docker-compose -f docker-compose.test.yaml up

gen:
  sqlc generate

create-migration name:
  migrate create -ext sql -dir ./db/migrations -seq -digits 4 {{name}}

migrate *args:
  migrate -path ./db/migrations -database $DATABASE_URL {{args}}
  just gen

migrate-prod *args:
  echo "Migrating $PROD_DATABASE_URL"
  migrate -path ./db/migrations -database $PROD_DATABASE_URL {{args}}

migrate-river:
  river migrate-up --database-url $DATABASE_URL

migrate-prod-river:
  river migrate-up --database-url $PROD_DATABASE_URL
