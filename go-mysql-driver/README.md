First run the normal test to setup the accounts and then run this test.

Example:

```
[dvaneeden@dve-carbon go-mysql-driver]$ ./run-test.sh 
Testing with default_authentication_plugin=mysql_native_password

github.com/go-sql-driver/mysql	v1.3.0
root@tcp(127.0.0.1:4000)/test	Ping error: this user requires mysql native password authentication.
nopw@tcp(127.0.0.1:4000)/	Ping error: this user requires mysql native password authentication.
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	Ping error: this authentication plugin is not supported

github.com/go-sql-driver/mysql	v1.4.1
root@tcp(127.0.0.1:4000)/test	OK
nopw@tcp(127.0.0.1:4000)/	OK
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	OK

github.com/go-sql-driver/mysql	v1.5.0
root@tcp(127.0.0.1:4000)/test	OK
nopw@tcp(127.0.0.1:4000)/	OK
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	OK

github.com/go-sql-driver/mysql	v1.6.0
root@tcp(127.0.0.1:4000)/test	OK
nopw@tcp(127.0.0.1:4000)/	OK
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	OK

Testing with default_authentication_plugin=caching_sha2_password

github.com/go-sql-driver/mysql	v1.3.0
root@tcp(127.0.0.1:4000)/test	Ping error: this user requires mysql native password authentication.
nopw@tcp(127.0.0.1:4000)/	Ping error: this user requires mysql native password authentication.
native:nat@tcp(127.0.0.1:4000)/	Ping error: this user requires mysql native password authentication.
native:nat@unix(/tmp/tidb.sock)/	Ping error: this user requires mysql native password authentication.
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	Ping error: this authentication plugin is not supported

github.com/go-sql-driver/mysql	v1.4.1
root@tcp(127.0.0.1:4000)/test	OK
nopw@tcp(127.0.0.1:4000)/	OK
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	OK

github.com/go-sql-driver/mysql	v1.5.0
root@tcp(127.0.0.1:4000)/test	OK
nopw@tcp(127.0.0.1:4000)/	OK
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	OK

github.com/go-sql-driver/mysql	v1.6.0
root@tcp(127.0.0.1:4000)/test	OK
nopw@tcp(127.0.0.1:4000)/	OK
native:nat@tcp(127.0.0.1:4000)/	OK
native:nat@unix(/tmp/tidb.sock)/	OK
sha2:sha@tcp(127.0.0.1:4000)/?tls=skip-verify	OK
```
