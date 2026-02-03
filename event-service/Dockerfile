FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./

COPY . .

RUN go mod download && go mod tidy

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init

RUN CGO_ENABLED=0 GOOS=linux go build -o event-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/event-service .
COPY --from=builder /app/docs ./docs

EXPOSE 8082

CMD ["./event-service"]
