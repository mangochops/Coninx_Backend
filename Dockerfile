# Build stage
FROM golang:1.24.1 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server .

# Final stage
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/server .

# Render expects apps to listen on $PORT
ENV PORT=10000
EXPOSE 10000

CMD ["./server"]
