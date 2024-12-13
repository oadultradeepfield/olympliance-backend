FROM golang:1.23 AS builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]