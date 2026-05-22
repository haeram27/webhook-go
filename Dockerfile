FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod ./
COPY . .

# Build the Go application with optimizations for a smaller binary size
# -ldflags="-s -w" removes symbol table and debug information, which reduces the binary size.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook-go -ldflags="-s -w"

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/webhook-go /app/webhook-go

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/app/webhook-go"]
