FROM golang:1.25.0-alpine

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum* ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]