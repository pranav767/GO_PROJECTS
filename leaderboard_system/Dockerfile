# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /leaderboard-server ./cmd/server/

# Stage 2: Runtime (scratch, zero OS attack surface)
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /leaderboard-server /leaderboard-server

USER 65534:65534

EXPOSE 8080 9090

ENTRYPOINT ["/leaderboard-server"]
