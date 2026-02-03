FROM golang:1.23-alpine AS builder

WORKDIR /app

# Устанавливаем swag для генерации Swagger документации
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Копируем go.mod для кеширования слоя
COPY go.mod ./

# Копируем исходники
COPY . .

# Обновляем зависимости на основе импортов
RUN go mod tidy

# Генерируем Swagger документацию
RUN swag init --parseDependency --parseInternal

# Собираем бинарник с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o gateway .

# Production образ
FROM alpine:3.19

# Устанавливаем ca-certificates для HTTPS запросов, timezone data и wget для healthcheck
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/gateway .

# Копируем swagger документацию
COPY --from=builder /app/docs ./docs

# Меняем владельца и права
RUN chown -R appuser:appuser /app

# Переключаемся на non-root пользователя
USER appuser

EXPOSE 8080

CMD ["./gateway"]