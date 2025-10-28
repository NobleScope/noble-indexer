-include .env
export $(shell sed 's/=.*//' .env)

lint:
	golangci-lint run

test:
	go test -p 8 -timeout 120s ./...

api-docs:
	cd cmd/api && swag init --parseDependency --parseInternal

.PHONY: lint test api-docs
