# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY static/ ./static/

# Build the application
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/svg-web-elements ./cmd/server

# Final stage
FROM alpine:latest

# Add non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/svg-web-elements /app/svg-web-elements
COPY --from=builder /app/static /app/static

# Set proper permissions
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Command to run
CMD ["/app/svg-web-elements"]