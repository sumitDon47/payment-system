FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY notification-service/go.mod notification-service/go.sum ./
RUN go mod download
COPY notification-service/ .
RUN go build -o notification-service .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/notification-service .
RUN chown appuser:appgroup /app/notification-service
USER appuser
CMD ["./notification-service"]
