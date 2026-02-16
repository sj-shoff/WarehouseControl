FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o warehouse-control ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/warehouse-control .
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

EXPOSE 8037

CMD ["./warehouse-control"]