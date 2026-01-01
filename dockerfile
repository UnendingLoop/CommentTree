FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o commentTree ./main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/commentTree .
COPY .env .
COPY internal/web /app/internal/web
COPY internal/migrations/0001_comments_create.up.sql /app/migrations/
EXPOSE 8080
CMD ["./commentTree"]
