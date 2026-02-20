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

# Expose port
EXPOSE 8081

# Air will be run via docker-compose command
CMD ["air", "-c", ".air.toml"]