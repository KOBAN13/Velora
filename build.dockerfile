FROM golang:1.26.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/app /app/app

COPY config.env .

EXPOSE 8080
CMD ["/app/app"]