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
./from_to_linux_amd64 -config=example_config.yaml
```

## Example config

```yaml
input:
  type: "postgres"
  dsn: "postgres://from-to-user:from-to-passw@localhost:5432/from-to-db?sslmode=disable"
  pollSeconds: 10
  tables:
    - from:
        name: "sales"
        keyColumn: "id"
      to:
        topic: "public_sales"
        partitions: 3
        replicationFactor: 1

output:
  type: "kafka"
  bootstrapServers: "localhost:9094"
```
