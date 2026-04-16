# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o simple-shop ./cmd/api

# Production stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Sao_Paulo

WORKDIR /app

COPY --from=builder /app/simple-shop .
COPY --from=builder /app/static ./static

EXPOSE 7000

CMD ["./simple-shop"]
