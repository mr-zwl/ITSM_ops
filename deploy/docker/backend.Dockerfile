FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY backend/go.mod ./backend/
WORKDIR /src/backend
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/itsm-ops-backend ./cmd/server

FROM alpine:3.21

RUN addgroup -S itsm && adduser -S itsm -G itsm
WORKDIR /app

COPY --from=builder /out/itsm-ops-backend /app/itsm-ops-backend

ENV APP_ENV=production \
    APP_PORT=8080 \
    APP_LOG_LEVEL=info

EXPOSE 8080
USER itsm

ENTRYPOINT ["/app/itsm-ops-backend"]
