.PHONY: up app-up infra api create-admin create-doorman docker-build test coverage mocks down linter

## Arquivo de variáveis usado localmente; para usar outro: make up ENV_FILE=.env
ENV_FILE ?= .env.test
export ENV_FILE

## Sobe MongoDB + emulador do Pub/Sub + api em primeiro plano (mesma topologia do Cloud Run)
up:
	docker compose -f ./docker-compose.yml up --build

## Mesmo que o up, em background
app-up:
	docker compose -f ./docker-compose.yml up --build -d

## Sobe só MongoDB + emulador do Pub/Sub, com o push apontando para a api rodando no host (make api)
infra:
	PUBSUB_PUSH_ENDPOINT=http://host.docker.internal:8081/internal/pubsub/notifications \
		docker compose -f ./docker-compose.yml up -d mongodb pubsub pubsub-init pubsub-ui firebase

## Roda a API localmente (com o $(ENV_FILE); veja .env.example)
api:
	go run ./cmd/api

## Cria um usuário admin no Firebase (local: no emulador): make create-admin EMAIL=... NAME=... PASSWORD=...
create-admin:
	go run ./cmd/create-admin -email="$(EMAIL)" -name="$(NAME)" -password="$(PASSWORD)"

## Cria um porteiro no Firebase (local: no emulador): make create-doorman EMAIL=... NAME=... PASSWORD=...
create-doorman:
	go run ./cmd/create-admin -role=doorman -email="$(EMAIL)" -name="$(NAME)" -password="$(PASSWORD)"

## Gera a imagem Docker da API
docker-build:
	docker build -t entregador-api .

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
