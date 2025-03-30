FROM golang:1.22

WORKDIR /app

# Устанавливаем зависимости
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    ca-certificates \
    tzdata

# Копируем файлы проекта
COPY . .

# Собираем приложение
ENV CGO_ENABLED=1
ENV GOOS=linux
RUN go mod download && \
    go mod tidy && \
    go build -ldflags="-s -w" -o /app/weveryone_bot

# Создаем директорию для данных
RUN mkdir -p /app/data

# Запускаем бота
CMD ["/app/weveryone_bot"]

