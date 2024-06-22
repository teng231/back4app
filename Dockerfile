# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Your Name <abc.email@example.com>"

# Set the Current Working Directory inside the container
WORKDIR /

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN go mod tidy

# Disable Go Modules
ENV GO111MODULE=off

# Build the Go app
RUN go build -o back4app .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./back4app"]