
# Сборка приложения
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# Компилируем для Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o web-todo .

#  Финальный образ
FROM ubuntu:latest

WORKDIR /app

# Копируем бинарник и веб-ресурсы
COPY --from=builder /app/web-todo .
COPY --from=builder /app/web ./web

# Порт, который будет слушать приложение
EXPOSE 7540

# Переменные окружения (можно переопределить при запуске)
ENV TODO_PORT=7540 \
    TODO_DBFILE=/data/scheduler.db \
    TODO_PASSWORD= \
    TODO_FULL_NEXT_DATE=true \
    TODO_SEARCH=false

# Точка монтирования для базы данных
VOLUME /data

CMD ["./web-todo"]