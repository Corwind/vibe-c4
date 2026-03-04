FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vibe-c4-api ./cmd/server/

FROM alpine:3.20

RUN apk --no-cache add ca-certificates git
COPY --from=builder /vibe-c4-api /vibe-c4-api

EXPOSE 8080

ENTRYPOINT ["/vibe-c4-api"]
