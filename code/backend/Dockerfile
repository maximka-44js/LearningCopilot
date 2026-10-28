# Step 1: Modules caching
FROM golang:1.26.10-alpine3.23 AS modules

COPY go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

# Step 2: Builder
FROM golang:1.26.10-alpine3.23 AS builder

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/migrate ./cmd/migrate

FROM busybox:1.37.0-musl AS healthcheck

# Step 3: Final
FROM scratch

COPY --from=healthcheck /bin/busybox /busybox
COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
COPY --from=builder /bin/worker /worker
COPY --from=builder /bin/migrate /migrate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65532:65532

CMD ["/app"]
