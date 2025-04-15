# Dockerfile
FROM golang:1.23.8 AS builder

WORKDIR /app

# Copiar arquivos go.mod e go.sum
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar o código fonte
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/entregador

# Imagem final
FROM alpine:latest

WORKDIR /root/

# Copiar o binário compilado
COPY --from=builder /app/main .

# Expor a porta que sua aplicação usa
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]