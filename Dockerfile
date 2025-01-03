# Use the official Go image (make sure the Go version matches your local version)
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first (to leverage Docker layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Create a directory for templates and data
RUN mkdir -p /app/templates /app/data

# Copy templates and data
COPY templates/ /app/templates/
COPY data/ /app/data/

# Build the Go app
RUN CGO_ENABLED=0 go build -o main ./cmd/server

# Expose the app's listening port
EXPOSE 8080

# Command to run the app
CMD ["./main"]
