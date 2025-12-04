# Use multi-stage build to keep the final image small
FROM golang:1.25-alpine AS builder

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o sequencer-reader ./cmd/sequencer-reader
RUN CGO_ENABLED=0 GOOS=linux go build -o sequencer-capture ./cmd/sequencer-capture

# Final stage: Use a minimal alpine image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh arbuser

# Set working directory
WORKDIR /app

# Copy the binaries from the builder stage
COPY --from=builder /app/sequencer-reader .
COPY --from=builder /app/sequencer-capture .

# Change ownership to the non-root user
RUN chown -R arbuser:arbuser /app
USER arbuser

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:8080/health || exit 1

EXPOSE 8080

CMD ["./sequencer-reader", "-rpc", "https://arb1.arbitrum.io/rpc"]
