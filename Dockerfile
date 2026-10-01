# Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app

# Copiar arquivos go.mod e go.sum
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar o código fonte
COPY . .

# Binário a compilar: entregador (consumer Kafka) ou api (HTTP)
ARG APP=entregador

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/${APP}

# Imagem final
FROM alpine:latest

WORKDIR /root/

# Copiar o binário compilado
COPY --from=builder /app/main .

# Configs dos consumers, passadas via -config
COPY --from=builder /app/config/subscriber/deployments ./config/subscriber/deployments

# Expor a porta que sua aplicação usa
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]