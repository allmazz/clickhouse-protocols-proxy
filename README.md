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
