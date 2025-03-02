# Use an official Go image as the build stage
FROM golang:1.24.0 AS build

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application with CGO disabled and statically linked
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Use a minimal base image for production
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the built binary from the build stage
COPY --from=build /app/main .

# Expose the application port
EXPOSE 7777

# Run the application
CMD ["./main"]
