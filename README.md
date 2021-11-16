Put multiple client versions in `~/opt/mysql/<version>/bin/mysql`. If you are using dbdeployer you might already have this.

Recommended set of versions to test with:

- 5.1.7x
- 5.7.x
- 8.0.25 or below
- 8.0.27 or newer

```
./tidb-server -config tidb.toml
```

`tidb.toml` contents:
```
socket = "/tmp/tidb.sock"

[security]
auto-tls=true
```
