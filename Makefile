.PHONY: up app-up up-kafka infra api worker docker-build-api docker-build-worker test coverage mocks down linter

## Sobe MongoDB + emulador do Pub/Sub + api em primeiro plano (mesma topologia do Cloud Run)
up:
	docker compose -f ./docker-compose.yml up --build

## Mesmo que o up, em background
app-up:
	docker compose -f ./docker-compose.yml up --build -d

## Sobe o modo Kafka: Kafka + Kafka UI + api + worker (MESSAGING_PROVIDER=kafka no .env)
up-kafka:
	docker compose -f ./docker-compose.yml --profile kafka up --build

## Sobe só MongoDB + emulador do Pub/Sub, com o push apontando para a api rodando no host (make api)
infra:
	PUBSUB_PUSH_ENDPOINT=http://host.docker.internal:8081/internal/pubsub/notifications \
		docker compose -f ./docker-compose.yml up -d mongodb pubsub pubsub-init

## Roda a API localmente (com o .env; veja .env.example)
api:
	go run ./cmd/api

## Roda o worker localmente (só no modo Kafka)
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
	docker compose -f ./docker-compose.yml --profile kafka down

## Roda o golangci-lint com as regras do .golangci.yml
linter:
	golangci-lint run ./...
