package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	_ "github.com/go-sql-driver/mysql"
)

var verboseFlag bool

type ClientCredential struct {
	user       string
	host       string
	password   string
	authPlugin string
}

type ClientTestSuite struct {
	users []ClientCredential
}

func NewClientTestSuite() *ClientTestSuite {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	cts := ClientTestSuite{}
	cts.users = []ClientCredential{
		{"nopw", "%", "", "mysql_native_password"},
		{"nat", "%", "nat", "mysql_native_password"},
		{"sha", "%", "sha", "caching_sha2_password"},
		{"sock", "%", user.Username, "auth_socket"},
	}
	return &cts
}

func (cts *ClientTestSuite) getClients() ([]string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	clients, err := filepath.Glob(user.HomeDir + "/opt/mysql/*/bin/mysql")
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (cts *ClientTestSuite) setupUsers(db *sql.DB) error {
	for _, user := range cts.users {
		_, err := db.Exec("DROP USER IF EXISTS '" + user.user + "'@'" + user.host + "'")
		if err != nil {
			return err
		}

		createUserStatement := "CREATE USER '" + user.user + "'@'" + user.host + "' IDENTIFIED WITH " + user.authPlugin
		if user.password != "" {
			if user.authPlugin == "auth_socket" {
				createUserStatement += " AS '" + user.password + "'"
			} else {
				createUserStatement += " BY '" + user.password + "'"
			}
		}
		_, err = db.Exec(createUserStatement)
		if err != nil {
			fmt.Printf("Failed to setup user %s: %s\n", user.user, err)
		}
	}

	return nil
}

func (cts *ClientTestSuite) testClient(client string) (okCount int, failCount int, err error) {
	connOpts := [][]string{
		[]string{"-h", "127.0.0.1", "-P", "4000"},
		[]string{"-S", "/tmp/tidb.sock"},
	}
	parts := strings.Split(client, "/")
	clientVer := semver.New(parts[len(parts)-3])

	for _, user := range cts.users {
		for _, connOpt := range connOpts {
			expectFailure := false

			// auth_socket requires a socket connection to function
			if user.authPlugin == "auth_socket" && connOpt[0] == "-h" {
				expectFailure = true
			}

			// MySQL 5.1.x doesn't support authentication plugins
			if user.authPlugin != "mysql_native_password" {
				if clientVer.LessThan(*semver.New("5.1.99")) {
					expectFailure = true
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			args := []string{"-u", user.user}
			args = append(args, connOpt...)
			if user.password != "" {
				args = append(args, "-p"+user.password)
			}
			args = append(args, "-e", "QUIT")
			cmd := exec.CommandContext(ctx, client, args...)
			output, _ := cmd.CombinedOutput()
			exitCode := cmd.ProcessState.ExitCode()

			if exitCode == 0 && !expectFailure {
				if verboseFlag {
					fmt.Printf("\U00002705\t")
					fmt.Printf("  Command '%s' returned %d.\n", cmd.String(), cmd.ProcessState.ExitCode())
				}
				okCount++
			} else if exitCode != 0 && expectFailure {
				if verboseFlag {
					fmt.Printf("\U00002714\U0000FE0F\t")
					fmt.Printf("  Command '%s' returned %d.\n", cmd.String(), cmd.ProcessState.ExitCode())
					println(string(output))
				}
				if user.authPlugin == "caching_sha2_password" && !strings.Contains(string(output),
					"ERROR 1251 (08004): Client does not support authentication protocol requested by server; consider upgrading MySQL client") {
					fmt.Printf("Unexpected output:\n")
					println(string(output))
				}
				okCount++
			} else {
				fmt.Printf("\U0000274C\t")
				fmt.Printf("  Command '%s' returned %d.\n", cmd.String(), cmd.ProcessState.ExitCode())
				println(string(output))
				failCount++
			}
		}
	}

	fmt.Printf("Testing %s: success=%d, failures=%d\n", client, okCount, failCount)

	return okCount, failCount, nil
}

func main() {
	flag.BoolVar(&verboseFlag, "verbose", false, "Verbose")
	flag.Parse()
	fmt.Println("TiDB Client Test")

	// Control connection. TLS support is required for testing caching_sha2_password
	// so set tls=skip-verify here to make sure TLS is available.
	dbAdmin, err := sql.Open("mysql", "root@tcp(127.0.0.1:4000)/?tls=skip-verify")
	if err != nil {
		panic(err)
	}

	var tidbVersion string
	err = dbAdmin.QueryRow("SELECT tidb_version()").Scan(&tidbVersion)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connected to:\n\t%s\n", strings.ReplaceAll(tidbVersion, "\n", "\n\t"))

	cts := NewClientTestSuite()
	err = cts.setupUsers(dbAdmin)
	if err != nil {
		panic(err)
	}

	clients, err := cts.getClients()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Clients:\n")
	for _, client := range clients {
		fmt.Printf("\t%s\n", client)
	}
	fmt.Printf("-----------------------------------------------\n")

	exitCode := 0
	authMethods := []string{"mysql_native_password", "caching_sha2_password"}
	for _, authMethod := range authMethods {
		fmt.Printf("Testing with %s as default authentication plugin\n", authMethod)
		_, err = dbAdmin.Exec("SET GLOBAL default_authentication_plugin='" + authMethod + "'")
		if err != nil {
			panic(err)
		}
		for _, client := range clients {
			_, failCount, err := cts.testClient(client)
			if err != nil {
				panic(err)
			}
			if failCount > 0 {
				exitCode = 1
			}
		}
	}

	defer dbAdmin.Close()
	os.Exit(exitCode)
}
