# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy module files first for layer caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static binary.
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /build/bin/server ./server

# Non-root user, read-only filesystem where possible.
USER nonroot

EXPOSE 8080

ENTRYPOINT ["/app/server"]
