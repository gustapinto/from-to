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
## Lua support

**FromTo** supports [Lua](https://www.lua.org/) scripting to create row mappers, an example mapper can be found at [example/mappers.lua](https://github.com/gustapinto/from-to/blob/main/example/mappers.lua). It uses the [yuin/gopher-lua](https://github.com/yuin/gopher-lua?tab=readme-ov-file#differences-between-lua-and-gopherlua) VM and preloads some of its libraries for improved DX.

### Preloaded libraries

- [cjoudrey/gluahttp](https://github.com/cjoudrey/gluahttp)
- [layeh.com/gopher-json](https://github.com/layeh/gopher-json)
