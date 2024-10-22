# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies

WORKDIR /app

# Copy the entire project
COPY . .

# Download dependencies and verify
RUN go mod download
RUN go mod verify

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -v -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy templates directory
COPY --from=builder /app/templates ./templates

# Copy the SQLite database file
COPY --from=builder /app/bookings.db .

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["./main"]
