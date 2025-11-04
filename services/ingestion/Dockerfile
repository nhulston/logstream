FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev librdkafka-dev pkgconf

WORKDIR /build

COPY proto/gen ../proto/gen/
COPY services/ingestion/go.mod services/ingestion/go.sum ./
RUN go mod download

COPY services/ingestion ./

RUN go build -tags musl -o /ingestion ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates librdkafka

WORKDIR /app

COPY --from=builder /ingestion .

EXPOSE 50051

CMD ["./ingestion"]
