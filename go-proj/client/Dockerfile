# Используем официальный образ Go
FROM golang:1.22-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod tidy

# Копируем все файлы проекта в контейнер
COPY . .

# Собираем приложение
RUN go build -o client_app client/main.go

# Запускаем клиент
CMD ["/app/client_app"]