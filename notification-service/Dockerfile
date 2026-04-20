FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notification-service .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/notification-service .
RUN chown appuser:appgroup /app/notification-service
USER appuser
CMD ["./notification-service"]
