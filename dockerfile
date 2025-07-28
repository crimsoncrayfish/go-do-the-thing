FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN ls -R /app

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main .

FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app

COPY --from=builder /app/main ./main
COPY --from=builder /app/static ./static/

EXPOSE 8080

CMD ["./main"]
