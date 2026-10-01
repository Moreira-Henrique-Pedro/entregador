.PHONY: up app-up api worker docker-build-api docker-build-worker test coverage mocks down linter

## Sobe infraestrutura + api + worker em primeiro plano
up:
	docker compose -f ./docker-compose.yml up --build

## Sobe infraestrutura + api + worker em background
app-up:
	docker compose -f ./docker-compose.yml up --build -d

## Roda a API HTTP localmente
api:
	go run ./cmd/api

## Roda o worker de notificações localmente
worker:
	go run ./cmd/worker -config=config/subscriber/deployments/delivery_subscriber_internal_commands.json

## Gera a imagem Docker da API HTTP
docker-build-api:
	docker build --build-arg APP=api -t entregador-api .

## Gera a imagem Docker do worker
docker-build-worker:
	docker build --build-arg APP=worker -t entregador-worker .

## rodar todos os testes unitários
test:
	go test -race -count=1 ./...

## cobertura de testes (ignora os mocks gerados) e relatório HTML em coverage.html
coverage:
	go test -count=1 -coverprofile=coverage.raw.out ./...
	grep -v "/mocks/" coverage.raw.out > coverage.out
	rm coverage.raw.out
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html

## gera os mocks das interfaces (config em .mockery.yaml)
mocks:
	mockery

down:
	docker compose -f ./docker-compose.yml down

## Roda o golangci-lint com as regras do .golangci.yml
linter:
	golangci-lint run ./...
