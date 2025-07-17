# Build stage
FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN mkdir -p /var/log && touch /var/log/telegram-bot.log && chmod 644 /var/log/telegram-bot.log
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/bot ./cmd/bot.go

# Runtime stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/bin/bot /app/bot
CMD ["/app/bot"]
