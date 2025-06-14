FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/main .

EXPOSE 3014

CMD ["./main"] 