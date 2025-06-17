FROM golang:1.23 AS builder

WORKDIR /app


COPY go.mod ./
COPY go.sum ./

RUN go mod download
RUN go mod verify   
COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .
COPY --from=builder /app/.env .   


EXPOSE 8080

CMD ["./main"]

