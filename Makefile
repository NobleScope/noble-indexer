-include .env
export $(shell sed 's/=.*//' .env)

indexer:
	go run ./cmd/indexer -c ./configs/dipdup.yml

api:
	go run ./cmd/api -c ./configs/dipdup.yml

lint:
	golangci-lint run

test:
	go test -p 8 -timeout 120s ./...

api-docs:
	cd cmd/api && swag init --parseDependency --parseInternal

.PHONY: lint test api-docs
