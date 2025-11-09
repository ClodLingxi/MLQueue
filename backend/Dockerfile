# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# China MainLand Pepositories
# RUN echo "https://mirrors.aliyun.com/alpine/v3.20/main" > /etc/apk/repositories && \
#    echo "https://mirrors.aliyun.com/alpine/v3.20/community" >> /etc/apk/repositories
# ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mlqueue .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/mlqueue .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./mlqueue"]
