# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /cicero ./cmd/app

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /cicero /app/cicero
COPY --from=builder /build/migrations /app/migrations

RUN adduser -D -g '' cicero
USER cicero

EXPOSE 8080

ENTRYPOINT ["/app/cicero"]
