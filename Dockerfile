# Use the official Golang image as the base image
FROM golang:1.22.2

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code to the working directory
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8080
EXPOSE 8080

# Define the command to run the application
CMD ["./main"]