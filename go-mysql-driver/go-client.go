package main

import (
	"database/sql"
	"fmt"
	"runtime/debug"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	v, ok := debug.ReadBuildInfo()
	if ok {
		for _, m := range v.Deps {
			fmt.Printf("%s\t%s\n", m.Path, m.Version)
		}
	}

	dsns := []string{
		"root@tcp(127.0.0.1:4000)/test",
		"nopw@tcp(127.0.0.1:4000)/",
		"native:nat@tcp(127.0.0.1:4000)/",
		"native:nat@unix(/tmp/tidb.sock)/",
		"sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify",
	}
	for _, dsn := range dsns {
		print(dsn, "\t")
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("Connect error: %s\n", err)
			continue
		}
		err = db.Ping()
		if err != nil {
			fmt.Printf("Ping error: %s\n", err)
			continue
		}
		db.Close()
		println("OK")
	}
	print("\n")
}
