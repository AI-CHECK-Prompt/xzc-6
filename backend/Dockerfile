FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o windpower-api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/windpower-api .
COPY config/config.yaml config/config.yaml

RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["./windpower-api"]
