# FromTo

## Building

1. Clone this repository
2. Open the repository on terminal and run `make build` to build the binaries for all platforms, or `make build/<platform>` to build for an specific platform. Supported platforms:
   - linux
   - windows
   - macos

### Building Pre Requisites

1. [Go 1.24+](https://go.dev/)
2. [GNU Make](https://www.gnu.org/software/make/) (Optional)

## Example usage

```bash
./from_to_linux_amd64 -config=./postgres_kafka_example_config.yaml
```

## Example config

```yaml
input:
  connector: "postgres"
  dsn: "postgres://from-to-user:from-to-passw@localhost:5432/from-to-db?sslmode=disable"
  pollSeconds: 10 # (Optional, default=30)
  tables:
    - from:
        name: "sales"
        keyColumn: "id" # (Optional, default="id")
      to:
        topic: "public_sales"
        partitions: 3 # (Optional, default=3)
        replicationFactor: 1 # (Optional, default=1)
      mapper: # (Optional, default=null)
        type: "lua"
        filePath: "./example/mappers.lua"
        function: "map_sales_event"

output:
  connector: "kafka"
  bootstrapServers: "localhost:9094"
```
