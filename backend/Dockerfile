# Build stage
FROM golang:1.24-alpine AS builder

RUN mkdir -p /home/app
WORKDIR /home/app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

RUN mkdir -p /home/app
WORKDIR /home/app

# Copy the binary from builder
COPY --from=builder /home/app/main .

EXPOSE 3000

CMD ["./main"]