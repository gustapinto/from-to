# From To

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gustapinto/from-to/main/docs/images/diagram-dark.jpg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gustapinto/from-to/main/docs/images/diagram-light.jpg">
  <img src="https://raw.githubusercontent.com/gustapinto/from-to/main/docs/images/diagram-light.jpg">
</picture>


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
  postgresConfig:
    dsn: "postgres://from-to-user:from-to-passw@localhost:5432/from-to-db?sslmode=disable"
    pollSeconds: 10
    tables:
      - "sales"

outputs:
  salesKafkaOutput:
    connector: "kafka"
    kafkaConfig:
      bootstrapServers:
        - "localhost:9094"
      topics:
        - name: "publicSales"
          partitions: 3 # (Optional, default=3)
          replicationFactor: 1 # (Optional, default=1)

mappers:
  salesMapper:
    type: "lua"
    luaConfig:
      filePath: "./example/mappers.lua"
      function: "map_sales_event_with_http"

channels:
  salesChannel:
    from: "sales"
    to: "publicSales"
    output: "salesKafkaOutput"
    mapper: "salesMapper" # (Optional)
```
## Lua support

**FromTo** supports [Lua](https://www.lua.org/) scripting to create row mappers, an example mapper can be found at [example/mappers.lua](https://github.com/gustapinto/from-to/blob/main/example/mappers.lua). It uses the [yuin/gopher-lua](https://github.com/yuin/gopher-lua?tab=readme-ov-file#differences-between-lua-and-gopherlua) VM and preloads some of its libraries for improved DX.

### Preloaded libraries

- [cjoudrey/gluahttp](https://github.com/cjoudrey/gluahttp)
- [layeh.com/gopher-json](https://github.com/layeh/gopher-json)
