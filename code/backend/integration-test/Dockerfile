FROM golang:1.26.10-alpine3.23 AS modules

COPY go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

FROM golang:1.26.10-alpine3.23

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]
