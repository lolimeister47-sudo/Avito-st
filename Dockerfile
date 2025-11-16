# ========= STAGE 1: builder =========
FROM golang:1.22 AS builder

WORKDIR /app

# сначала модули, чтобы кешировать зависимости
COPY go.mod go.sum ./
RUN go mod download

# остальной код
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pr-service ./cmd/pr-service

# ========= STAGE 2: runtime =========
FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/pr-service /app/pr-service

ENV HTTP_ADDR=":8080"

USER app

EXPOSE 8080

ENTRYPOINT ["/app/pr-service"]
