FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g /cmd/main.go --parseDependency --parseInternal --output ./api/docs
RUN go build -o subscription-aggregator ./cmd
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/subscription-aggregator .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/api/docs /app/docs
COPY internal/db/migrations ./internal/db/migrations

RUN apk add --no-cache bash postgresql-client

ENTRYPOINT ["./subscription-aggregator"]