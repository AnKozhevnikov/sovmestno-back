# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go.mod
COPY go.mod ./

# Copy source code
COPY . .

# Update dependencies based on imports
RUN go mod tidy

# Generate swagger docs
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o application-service .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/application-service .
COPY --from=builder /app/docs ./docs

EXPOSE 8083

CMD ["./application-service"]
