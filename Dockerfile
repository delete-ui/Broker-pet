FROM golang:1.24.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/server

FROM alpine:latest

RUN apk add --no-cache tzdata

COPY --from=builder /app/main /main

CMD ["/main"]