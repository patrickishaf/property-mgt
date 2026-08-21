.PHONY: run build build-macos build-linux demo

run:
	go run cmd/server/main.go

run-web:
	npm --prefix ./web run dev

build:
	go build -o ./bin/app cmd/server/main.go

build-macos:
	go build -o ./bin/macos/app cmd/server/main.go

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./bin/linux/app cmd/server/main.go

demo:
	go run cmd/test/main.go

migrate:
	go run cmd/migration/main.go

migrate-down:
	go run cmd/migration/main.go down

migrate-cleanup:
	migrate -path ./internal/db/migrations -database "postgres://peterndukwe@localhost:5432/hookupxx?sslmode=disable" force 1

migrate-reset:
	go run cmd/migration/main.go reset
