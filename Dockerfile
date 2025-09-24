# ===== build stage =====
FROM golang:1.24 AS build
WORKDIR /app

# зависимости
COPY go.mod go.sum ./
RUN go mod download

# весь код
COPY . .

# собираем статически (modernc sqlite — pure Go, CGO не нужен)
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -o /app/todo .

# ===== runtime stage =====
FROM ubuntu:latest
WORKDIR /app

# бинарник и фронтенд
COPY --from=build /app/todo /usr/local/bin/todo
COPY web ./web

# дефолты (можно переопределить при запуске)
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
# ENV TODO_PASSWORD=   # при необходимости задавай на run

EXPOSE 7540
CMD ["/usr/local/bin/todo"]
