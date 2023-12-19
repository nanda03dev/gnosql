# Use the official Go image as the base
FROM golang:latest

ENV GIN_MODE=release

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go server
RUN go build -o gnosql .

# # Expose the port that the server listens on

EXPOSE 5454

# Set the command to run the server when the container starts
CMD ["./gnosql"]
