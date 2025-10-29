# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build argument to specify which service to build
ARG SERVICE_NAME
RUN if [ -z "$SERVICE_NAME" ]; then echo "SERVICE_NAME build arg is required" && exit 1; fi

# Build the specified service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./cmd/${SERVICE_NAME}/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/service /app/service

# Copy migration files
ARG SERVICE_NAME
COPY --chown=appuser:appuser db/ /app/db/

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (default 8080, can be overridden)
EXPOSE 8080

# Run the service
CMD ["/app/service"]
