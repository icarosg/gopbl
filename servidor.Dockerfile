FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o servidor ./servidor/servidor.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/servidor .
COPY --from=builder /app/modelo ./modelo
COPY --from=builder /app/servidor/postos.json ./postos.json
COPY --from=builder /app/servidor/veiculos.json ./veiculos.json

EXPOSE 8080
CMD ["./servidor"]
