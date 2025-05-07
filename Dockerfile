# Используем образ golang для сборки приложения
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и устанавливаем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Финальный контейнер для запуска приложения
FROM alpine:latest

WORKDIR /app

# Копируем собранное приложение из образа сборки
COPY --from=builder /app/main .

# Копируем файл с переменными окружения
COPY config.env ./

# Устанавливаем необходимые пакеты для работы с PostgreSQL и bash
RUN apk add --no-cache bash postgresql-client

# Открываем порт 8080
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
