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

# Next steps

For authentication issues it can be useful to capture packets with something like `tshark -i lo -w /tmp/tidb.pcap port 4000` and later inspect them with Wireshark and compare this with MySQL. While Wireshark has [the ability to decrypt TLS traffic](https://wiki.wireshark.org/TLS#tls-decryption) you probably should disable SSL/TLS if you can. 

In wireshark you probably want to use "Decode As..." to configure Wireshark to decode traffic on port 4000 with the MySQL protocol dissector.

Protocol documentation can be found [here](https://dev.mysql.com/doc/dev/mysql-server/latest/PAGE_PROTOCOL.html).
