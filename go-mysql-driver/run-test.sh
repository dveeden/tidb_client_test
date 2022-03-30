#!/bin/bash

for auth in mysql_native_password caching_sha2_password
do
  echo -e "Testing with default_authentication_plugin=${auth}\n"
  mysql -h 127.0.0.1 -u root -P 4000 -e "SET GLOBAL default_authentication_plugin='${auth}'"

  for v in 1.3.0 1.4.1 1.5.0 1.6.0
  do
    go mod edit -require github.com/go-sql-driver/mysql@v${v}
    go mod download github.com/go-sql-driver/mysql
    go build && ./go-mysql-driver $*
  done

done
