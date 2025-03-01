# Use an official Go image as the build stage
FROM golang:1.24.0 AS build

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application (Set correct path)
RUN go build -o main ./cmd/main.go

# Use a minimal base image for production
FROM alpine:latest

# Install necessary dependencies (e.g., libc)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the built binary from the build stage
COPY --from=build /app/main .

# Expose the application port
EXPOSE 7777

# Run the application
CMD ["./main"]
