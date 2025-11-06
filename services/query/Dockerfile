FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY services/query/go.mod services/query/go.sum ./
RUN go mod download

COPY services/query ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /query ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /query .

EXPOSE 8080

CMD ["./query"]
