# From To

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gustapinto/from-to/main/docs/images/diagram-dark.jpg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gustapinto/from-to/main/docs/images/diagram-light.jpg">
  <img src="https://raw.githubusercontent.com/gustapinto/from-to/main/docs/images/diagram-light.jpg">
</picture>

## Installation

The pre-built binaries for **FromTo** are available in the [releases page](https://github.com/gustapinto/from-to/releases), they are the recommended installation option, but you can also compile the code yourself

### Compiling as binaries

1. To compile **FromTo** first clone this repository in your machine
2. Then open the cloned folder on a shell and run `make build` to build the binaries for all supported platforms, or run `make build/<platform>` to build for an specific platform. The currently supported platforms are:
   - linux
   - windows
   - macos

#### Building Pre Requisites

1. [Go 1.24+](https://go.dev/)
2. [GNU Make](https://www.gnu.org/software/make/) (Optional)

## Example usage

```bash
./from_to_linux_amd64 -manifest=./from_to.yaml
```

## Example config

```yaml
input:
  connector: "postgres"
  postgresConfig:
    dsn: "postgres://from-to-user:from-to-passw@localhost:5432/from-to-db?sslmode=disable"
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
    mapper: "salesMapper"
```

Additional options and specific documentation on inputs, outputs and mappers can be found at the [examples folder](https://github.com/gustapinto/from-to/blob/main/example)

## Connector support

The currently supported connectors and mappers are:

- **PostgreSQL (postgres):** Input connector
- **Kafka (kafka):** Output connector
- **Webhook (webhook):** Output connector
- **Lua (lua):** Mapper

## Lua support

**FromTo** supports [Lua](https://www.lua.org/) scripting to create row mappers, an example mapper can be found at [example/mappers.lua](https://github.com/gustapinto/from-to/blob/main/example/mappers.lua). It uses the [yuin/gopher-lua](https://github.com/yuin/gopher-lua) VM and preloads some of its libraries for improved DX.

### Preloaded libraries

- [cjoudrey/gluahttp](https://github.com/cjoudrey/gluahttp)
- [layeh.com/gopher-json](https://github.com/layeh/gopher-json)

## Development

**FormTo** development follows the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

### Development Pre Requisites

1. [Go 1.24+](https://go.dev/)
2. [GNU Make](https://www.gnu.org/software/make/) (Optional)
3. [Docker](https://www.docker.com/) or an compatible alternative
4. [Docker Compose](https://docs.docker.com/compose/) or an compatible alternative
