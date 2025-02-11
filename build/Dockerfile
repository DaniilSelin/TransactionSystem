# Используем официальный образ Go для сборки
FROM golang:1.22 AS builder

WORKDIR /app

# Копируем модули и зависимости
COPY ../go.mod ../go.sum ./
RUN go mod download

# Копируем весь исходный код (учитывая, что мы в ./build)
COPY ../ ./

# Сборка бинарного файла
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o transaction-system ./cmd/server/main.go

# Используем минимальный образ для финального контейнера
FROM alpine:latest

WORKDIR /root/

# Устанавливаем зависимости
RUN apk --no-cache add ca-certificates

# Копируем собранный бинарник
COPY --from=builder /app/transaction-system .

# Копируем конфигурацию
COPY ../config/config.yml ./config/config.yml

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./transaction-system"]
