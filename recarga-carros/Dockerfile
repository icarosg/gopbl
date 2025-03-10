FROM golang:1.18 AS build

WORKDIR /src

# copiado para dentro de src
COPY go.mod ./
COPY main.go ./

RUN go build -o /server

# EXPOSE 8080

# # passando o que será executado, a partir do build anteriormente
# CMD [ "/server" ]

# usando distroless para deixar a imagem bem menor

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "/server" ]