package main

import (
	"database/sql"
	"flag"
	"fmt"
	"runtime/debug"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := flag.String("host", "127.0.0.1", "database host")
	port := flag.Int("port", 4000, "database port")
	flag.Parse()
	v, ok := debug.ReadBuildInfo()
	if ok {
		for _, m := range v.Deps {
			fmt.Printf("%s\t%s\n", m.Path, m.Version)
		}
	}

	dsns := []string{
		fmt.Sprintf("root@tcp(%s:%d)/test?timeout=10s", *host, *port),
		fmt.Sprintf("nopw@tcp(%s:%d)/", *host, *port),
		fmt.Sprintf("native:nat@tcp(%s:%d)/", *host, *port),
		"native:nat@unix(/tmp/tidb.sock)/",
		fmt.Sprintf("sha2:sha@tcp(%s:%d)/?tls=skip-verify", *host, *port),
	}
	for _, dsn := range dsns {
		print(dsn, "\t")

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("Connect error: %s\n", err)
			continue
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			fmt.Printf("Ping error: %s\n", err)
			continue
		}

		println("OK")
	}
	print("\n")
}
