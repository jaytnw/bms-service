# FROM golang:1.23-alpine

# WORKDIR /app
# COPY . .

# RUN go build -o bms-service ./cmd/server

# CMD ["./bms-service"]

# Stage 1: build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o bms-service ./cmd/server

# Stage 2: run (clean, minimal image)
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bms-service .

# Set binary as entrypoint
CMD ["./bms-service"]
