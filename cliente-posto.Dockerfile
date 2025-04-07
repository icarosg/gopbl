FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o cliente-posto ./cliente-posto/cliente-posto.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/cliente-posto .
COPY --from=builder /app/modelo ./modelo

EXPOSE 8081
#EXPOSE 22
CMD ["./cliente-posto"] 