FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /vestigo ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /vestigo /usr/local/bin/vestigo
COPY migrations /migrations
EXPOSE 8080
ENTRYPOINT ["vestigo"]
