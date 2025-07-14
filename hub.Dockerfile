# Build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY gen/ ./gen/
COPY cmd/ ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hub ./cmd/hub/main.go

# Move to lightweight image
FROM scratch

COPY --from=builder /app/hub /hub

ENTRYPOINT ["/hub"]