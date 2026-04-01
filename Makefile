.PHONY: run build test test-cover vet docker clean

run:
	go run ./cmd

build:
	CGO_ENABLED=0 go build -o jarvis ./cmd

test:
	go test -race ./...

test-cover:
	go test -race -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

vet:
	go vet ./...

docker:
	docker compose up -d --build jarvis

docker-all:
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -f jarvis cmd.exe coverage.out
