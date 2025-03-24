FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY ../go.mod ./
RUN go mod download

COPY ./cliente ./cliente


RUN go build -o cliente ./cliente/cliente.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/cliente .
COPY --from=builder /app/modelo ./modelo  # Inclui a pasta modelo na imagem final

EXPOSE 8081
CMD ["./cliente"]
