sqlc: # generated db code
	sqlc generate

generate:
	go generate ./...

up: # docker-compose up
	docker-compose -f docker-compose.yml up -d --build

up_log: # docker-compose up
	docker-compose -f docker-compose.yml up --build

down: ## docker-compose down
	docker-compose -f docker-compose.yml down --volumes

test: ## test: run unit test
	go test -v -race -cover -coverprofile coverage.txt -covermode=atomic ./...

.PHONY: sqlc generate up up_log down test