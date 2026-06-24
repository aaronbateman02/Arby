.PHONY: build test vet tidy docker-build docker-up docker-down

build:
	go build -o polybot ./cmd/polybot

test:
	go test ./... -v -count=1

vet:
	go vet ./...

tidy:
	go mod tidy

docker-build:
	docker compose build

docker-up:
	docker compose -p arby up -d

docker-down:
	docker compose -p arby down

docker-logs:
	docker compose -p arby logs -f
