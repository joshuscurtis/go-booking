# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the entire project
COPY . .

# Generate templ files
RUN templ generate

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -v -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache libc6-compat

WORKDIR /root/

# Copy the binary and assets
COPY --from=builder /app/main .
COPY --from=builder /app/bookings.db .

EXPOSE 8080

CMD ["./main"]
