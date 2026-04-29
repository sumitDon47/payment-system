FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && go build -o payment-service .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/payment-service .
RUN chown appuser:appgroup /app/payment-service
USER appuser
EXPOSE 9090
CMD ["./payment-service"]
