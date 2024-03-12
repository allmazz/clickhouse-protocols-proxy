# clickhouse-protocol-proxy
This proxy accepts ClickHouse HTTP requests and forwards them to a specified ClickHouse server using the native protocol.

## Why you need it
You don't need it. It was developed as a temporary hotfix due to firewall issues.

## Auth
The proxy creates a connection with credentials specified in the HTTP request and maintains it until stopped.

## Details
- Query parameters are supported.
- Non-read-only queries received from GET requests are blocked, as ClickHouse does.
- The `statistics` and `rows_before_limit_at_least` fields in responses always have zero values.

## Configuration example
```
log:
  level: info
target:
  hosts:
    - "clickhouse-server:9000"
  settings:
    max_execution_time: 60

  maxConnectionPerUser: 5
  maxConnectionLifetime: 10m
  dialTimeout: 1s
  readTimeout: 5m

  debug: false
server:
  addr: "127.0.0.1:8123"
```

## Build and run
- `go mod download`
- `go build -v -o clickhouse-protocol-proxy cmd/clickhouse-protocol-proxy/main.go`
- `./clickhouse-protocol-proxy config.yaml`

## Deploy with docker
`docker run -it -e TARGET_HOST=clickhouse-server:9000 allmazz/clickhouse-protocol-proxy:latest`
#### Configuration with environment variables
```
LOG_LEVEL=info
TARGET_HOST=clickhouse-server:9000
TARGET_MAX_CONNECTION_PER_USER=5
TARGET_MAX_CONNECTION_LIFETIME=10m
TARGET_DIAL_TIMEOUT=1s
TARGET_READ_TIMEOUT=5m
TARGET_DEBUG=false
SERVER_ADDR=0.0.0.0:8123
```
