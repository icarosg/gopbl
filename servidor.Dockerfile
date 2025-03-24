FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY ../go.mod ./
RUN go mod download

COPY ./servidor ./servidor


RUN go build -o servidor ./servidor/servidor.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/servidor .
COPY --from=builder /app/modelo ./modelo  # Inclui a pasta modelo na imagem final

EXPOSE 8080
CMD ["./servidor"]
