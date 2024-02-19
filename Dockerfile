FROM golang:1.22 AS builder

WORKDIR /app

COPY ./api ./

RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/myapp ./
COPY --from=builder /app/config.toml ./

CMD ["./myapp"]
