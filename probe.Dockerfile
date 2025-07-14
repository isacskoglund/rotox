# Build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY gen/ ./gen/
COPY cmd/ ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o probe ./cmd/probe/main.go

# Move to lightweight image
FROM scratch

COPY --from=builder /app/probe /probe

ENTRYPOINT ["/probe"]