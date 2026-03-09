FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.21
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /bin/api /app/api
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/migrations /app/migrations
EXPOSE 8080
USER appuser
CMD ["/app/api"]
