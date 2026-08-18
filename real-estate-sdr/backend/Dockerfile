# ==========================================
# Stage 1 - Build
# ==========================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Dependências primeiro para aproveitar o cache
COPY go.mod go.sum ./

RUN go mod download

# Código da aplicação
COPY . .

# Compila o binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o api ./cmd/api


# ==========================================
# Stage 2 - Runtime
# ==========================================
FROM alpine:3.22

WORKDIR /app

# Copia apenas o binário
COPY --from=builder /app/api ./api

EXPOSE 8080

ENTRYPOINT ["./api"]