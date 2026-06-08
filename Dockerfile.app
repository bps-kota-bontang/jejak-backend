FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o jejak-backend ./cmd/app

FROM alpine:3.22

ARG BUILD_HASH
ENV APP_BUILD=${BUILD_HASH}

RUN apk add --no-cache ca-certificates && \
    adduser -D -g '' jejak && \
    mkdir -p /app/public/geojson && \
    chown -R jejak:jejak /app/public

WORKDIR /app

COPY --from=builder /app/jejak-backend .

USER jejak

CMD ["./jejak-backend"]
