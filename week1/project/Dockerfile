# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/hello-ci .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/hello-ci .
CMD ["/app/hello-ci"]
