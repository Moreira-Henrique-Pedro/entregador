# Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app

# Copiar arquivos go.mod e go.sum
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar o código fonte
COPY . .

# Binário a compilar: api (HTTP + push do Pub/Sub) ou worker (só no modo Kafka)
ARG APP=api

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/${APP}

# Imagem final
FROM alpine:latest

# Certificados para TLS (MongoDB Atlas, Twilio, Google APIs) e timezone
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app

WORKDIR /app

# Copiar o binário compilado
COPY --from=builder /app/main .

# Config do consumer do worker, passada via -config
COPY --from=builder /app/config/subscriber/deployments ./config/subscriber/deployments

USER app

# O Cloud Run injeta PORT (padrão 8080); localmente vale HTTP_PORT (padrão 8081)
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]