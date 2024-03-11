FROM golang:latest as build
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -v -o clickhouse-protocol-proxy cmd/clickhouse-protocol-proxy/main.go


FROM alpine:latest
RUN apk update && apk add gettext && apk cache clean
WORKDIR /clickhouse-protocol-proxy

COPY --from=build /build/clickhouse-protocol-proxy /build/config.template.yaml /build/entrypoint.sh ./
RUN chmod +x entrypoint.sh clickhouse-protocol-proxy

ENV LOG_LEVEL info
ENV TARGET_HOST clickhouse-server:9000
ENV TARGET_MAX_CONNECTION_PER_USER 5
ENV TARGET_MAX_CONNECTION_LIFETIME 10m
ENV TARGET_DIAL_TIMEOUT 1s
ENV TARGET_READ_TIMEOUT 5m
ENV TARGET_DEBUG false
ENV SERVER_ADDR 0.0.0.0:8123

ENTRYPOINT ["/bin/sh", "entrypoint.sh"]
