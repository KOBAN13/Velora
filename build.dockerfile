FROM golang:1.26.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY server ./server
RUN rm -f server/Internal/server/match/nutrient_spawner.go
COPY shared ./shared
COPY esc ./esc
COPY systems ./systems
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/app /app/app

COPY config.env .

EXPOSE 8080
CMD ["/app/app"]
