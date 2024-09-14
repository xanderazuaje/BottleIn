# backend/Dockerfile

# Build Stage
FROM golang:1.23-alpine AS builder

# Install git, swag CLI, and necessary packages
RUN apk update && apk add --no-cache git && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Set environment variables
ENV GO111MODULE=on

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Generate Swagger docs
RUN swag init -g cmd/server/main.go

# Build the Go application
RUN go build -o server cmd/server/main.go

# Run Stage
FROM alpine:latest

# Install ca-certificates
RUN apk update && apk add --no-cache ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary and docs from builder
COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

# Set environment variables
ENV PORT=8080
ENV MONGO_URI=mongodb://mongo:27017/bottlenet

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
