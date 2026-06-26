# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /smartpost ./cmd/smartpost

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata ffmpeg

WORKDIR /app

COPY --from=builder /smartpost /app/smartpost
COPY internal/database/migrations/ /app/internal/database/migrations/

EXPOSE 8080

CMD ["/app/smartpost"]
