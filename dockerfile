FROM golang:1.24.1-alpine AS builder

RUN apk add --no-cache nodejs npm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

RUN templ generate

RUN npm install htmx.org
RUN mv node_modules/htmx.org/dist/htmx.min.js static/js/htmx.min.js

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main .

FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app

COPY --from=builder /app/main ./main
COPY --from=builder /app/static ./static/

EXPOSE 8080

CMD ["./main"]
