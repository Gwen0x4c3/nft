# Multi-stage Dockerfile for NFT Platform

# Base image with Go
FROM golang:1.21-alpine AS base
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app

# Development stage
FROM base AS development
RUN apk add --no-cache make gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["go", "run", "./cmd/server"]

# Build stage
FROM base AS builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpc-server ./cmd/grpc-server

# Production stage
FROM alpine:3.18 AS production
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/grpc-server .
COPY --from=builder /app/migrations ./migrations/
EXPOSE 8080 9000
CMD ["./server"]