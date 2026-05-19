FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY collector/go.mod ./collector/
WORKDIR /src/collector
RUN go mod download

COPY collector/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/itsm-collector ./cmd/agent

FROM alpine:3.21

RUN addgroup -S itsm && adduser -S itsm -G itsm
WORKDIR /app

COPY --from=builder /out/itsm-collector /app/itsm-collector

ENV COLLECTOR_ENDPOINT=http://backend:8080/api/v1/collect \
    COLLECTOR_INTERVAL=60 \
    COLLECTOR_MODE=auto

USER itsm

ENTRYPOINT ["/app/itsm-collector"]
