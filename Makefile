.PHONY: up app-up api docker-build-api test down linter

## Inicializa apenas o docker-compose
up:
	docker compose -f ./docker-compose.yml up --build

## Inicializa toda a aplicação
app-up:
	docker compose -f ./docker-compose.yml up --build -d
	./scripts/run/run.application.sh

## Roda a API HTTP de consulta localmente
api:
	go run ./cmd/api

## Gera a imagem Docker da API HTTP
docker-build-api:
	docker build --build-arg APP=api -t entregador-api .

## rodar todos os testes unitários
test:
	go test -v -coverprofile=coverage.out ./internal/...

down:
	docker compose -f ./docker-compose.yml down

## Roda o golangci-lint com as regras do .golangci.yml
linter:
	golangci-lint run ./...
