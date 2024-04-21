# Base image
FROM golang:1.20
# Set the working directory
WORKDIR /app
# Copy go.mod and go.sum
COPY go.mod go.sum ./
# Download Go module dependencies
RUN go mod download
# Copy the rest of the application files
COPY . .
# Expose the port(s) required by your application
EXPOSE 8080
# Define environment variables, if needed
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USERNAME=root
ENV DB_PASSWORD=secret
ENV DB_NAME=mydatabase
# Set up the database (You can modify this based on your database setup)
# Set up SQLite
RUN apt-get update && apt-get install -y sqlite3
# Build the Go application
RUN go build -o forum
# Start the application
CMD ["./forum"]