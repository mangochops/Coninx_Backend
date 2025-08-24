# Start from official Go image
FROM golang:1.22 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o server .

# Final image
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/server .

# Expose Render's expected port
EXPOSE 10000
CMD ["./server"]
