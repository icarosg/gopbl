FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o cliente ./cliente-veiculo/cliente.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/cliente .
COPY --from=builder /app/modelo ./modelo

EXPOSE 8082
CMD ["./cliente"] 