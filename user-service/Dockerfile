FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && go build -o user-service .

# Final stage — tiny image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/user-service .
EXPOSE 8080
CMD ["./user-service"]
