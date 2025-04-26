FROM golang:1.24-alpine as builder

# Install build dependencies
RUN apk add --no-cache make git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application (with CGO disabled for static linking)
RUN CGO_ENABLED=0 make

# Use distroless as minimal base image
FROM gcr.io/distroless/static

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/mcp-mysql /app/bin/mcp-mysql
COPY --from=builder /app/config.yml /app/config.yml

# Define environment variables with default values
ENV LOG_PATH=""
ENV DEBUG="false"
ENV MYSQL_HOST="localhost"
ENV MYSQL_PORT="3306"
ENV MYSQL_USER="root"
ENV MYSQL_PASSWORD=""
ENV MYSQL_DATABASE=""
ENV MYSQL_DSN=""
ENV MYSQL_READ_ONLY="false"
ENV MYSQL_EXPLAIN_CHECK="false"

# Set entrypoint
ENTRYPOINT ["/app/bin/mcp-mysql", "server", "--config=/app/config.yml"]
