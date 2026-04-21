# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o analytics-service .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/analytics-service .

EXPOSE 8084

CMD ["./analytics-service"]
