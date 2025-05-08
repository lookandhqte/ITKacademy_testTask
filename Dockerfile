FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main ./

COPY config.env ./config.env

RUN apk add --no-cache bash postgresql-client

EXPOSE 8080

CMD ["./main"]
