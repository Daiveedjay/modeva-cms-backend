FROM golang:1.25-alpine

# Install build dependencies
RUN apk add --no-cache git make bash curl

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Copy startup script
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Expose port
EXPOSE 8081

# Run migrations and start app
CMD ["/app/start.sh"]