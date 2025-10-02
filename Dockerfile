FROM golang:1.25.1-alpine

WORKDIR /app

# Устанавливаем зависимости
RUN apk add --no-cache git

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/app

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]