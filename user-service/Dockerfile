FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && go build -o user-service .

# Final stage — tiny image
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/user-service .
RUN chown appuser:appgroup /app/user-service
USER appuser
EXPOSE 8080
CMD ["./user-service"]
