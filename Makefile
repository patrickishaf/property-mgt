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
