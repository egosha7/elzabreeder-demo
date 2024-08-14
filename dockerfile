# Используем официальный образ Golang как базовый
FROM golang:1.22-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY server/go.mod server/go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем остальной исходный код
COPY server .

# Переходим в директорию с main.go
WORKDIR /app/cmd

# Собираем приложение
RUN go build -o /myapp

# Копируем шаблоны и статические файлы
WORKDIR /app
COPY server/cmd/templates ./cmd/templates
COPY server/cmd/static ./cmd/static

# Используем минималистичный образ для запуска нашего приложения
FROM alpine:latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем скомпилированное приложение из предыдущего этапа
COPY --from=builder /myapp /myapp
COPY --from=builder /app/cmd/templates ./cmd/templates
COPY --from=builder /app/cmd/static ./cmd/static

# Копируем .aws папку для AWS клиента
COPY .aws /root/.aws

EXPOSE 443

# Указываем команду запуска приложения
ENTRYPOINT ["/myapp"]