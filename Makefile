.PHONY: up app-up

## Inicializa apenas o docker-compose
up:
	docker compose -f ./docker-compose.yml up --build

## Inicializa toda a aplicação
app-up:
	docker compose -f ./docker-compose.yml up --build -d
	./scripts/run/run.application.sh

## rodar todos os testes unitários
test:
	go test -v -coverprofile=coverage.out ./internal/...