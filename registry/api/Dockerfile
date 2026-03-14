# Stage 1: builder
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache ca-certificates git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /registry ./cmd/registry

# Stage 2: minimal runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /registry /registry
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT ["/registry"]
