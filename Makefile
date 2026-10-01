.PHONY: up app-up api docker-build-api test coverage mocks down linter

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
