.PHONY: build test run generate migrate-up migrate-down docker-up docker-down clean

build:
	go build -o vestigo ./cmd/server

test:
	go test ./... -v -count=1

run: build
	./vestigo

generate:
	go run github.com/99designs/gqlgen generate

migrate-up:
	migrate -path migrations -database "postgres://vestigo:vestigo@localhost:5432/vestigo?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://vestigo:vestigo@localhost:5432/vestigo?sslmode=disable" down

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

clean:
	rm -f vestigo
	docker compose down -v
