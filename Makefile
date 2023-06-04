export CONTAINER_NAME=gophermart-db
export DB_USERNAME=postgres
export DB_PASSWORD=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=gophermart

export RUN_ADDRESS=:8080
export DATABASE_URI=host=localhost port=5432 user=postgres password=postgres dbname=gophermart sslmode=disable

start-db:
	docker start $(CONTAINER_NAME) || docker run -d --name $(CONTAINER_NAME) -p $(DB_PORT):$(DB_PORT) \
		-e POSTGRES_USER=$(DB_USERNAME) -e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) postgres

build-app:
	cd cmd/gophermart && go build -o gophermart

run-app:
	cd cmd/gophermart && ./gophermart

start: start-db build-app run-app

stop-db:
	docker stop $(CONTAINER_NAME)
	docker rm $(CONTAINER_NAME)

stop: stop-db

lint:
	go vet ./...
	golangci-lint run ./...

.PHONY: run-app start-db start stop-db stop lint
