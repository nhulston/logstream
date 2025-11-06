FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev librdkafka-dev pkgconf

WORKDIR /build

COPY proto/gen ../proto/gen/
COPY services/consumer/go.mod services/consumer/go.sum ./
RUN go mod download

COPY services/consumer ./

RUN go build -tags musl -o /consumer ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates librdkafka

WORKDIR /app

COPY --from=builder /consumer .

CMD ["./consumer"]
