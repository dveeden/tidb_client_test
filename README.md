# Usage

Put multiple client versions in `~/opt/mysql/<version>/bin/mysql`. If you are using [dbdeployer](https://github.com/datacharmer/dbdeployer) you might already have this.

Recommended set of versions to test with:

- 5.1.7x
- 5.7.x
- 8.0.25 or below
- 8.0.27 or newer

Run a server like this:
```
./tidb-server -config tidb.toml
```

`tidb.toml` contents:
```
socket = "/tmp/tidb.sock"

[security]
auto-tls=true
```

Then build this tool with `go build` and run it like `./tidb_client_test`. Add `-verbose` to get more verbose output.
