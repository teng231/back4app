# Stage 1: Build stage
FROM public.ecr.aws/docker/library/golang:1.22-alpine3.20 AS build

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o back4app .

# Stage 2: Final stage
FROM public.ecr.aws/docker/library/alpine:3.18

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/back4app .

# Set the timezone and install CA certificates
RUN apk --no-cache add ca-certificates tzdata curl nano

# Set the entrypoint command
ENTRYPOINT ["/app/back4app", "start"]