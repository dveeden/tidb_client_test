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
var version string

type clientCredential struct {
	user       string
	host       string
	password   string
	authPlugin string
}

type clientTestSuite struct {
	users []clientCredential
}

type clientTestResults struct {
	okCount     int
	failCount   int
	testResults []clientTestResult
}

type clientTestResult struct {
	user          clientCredential
	client        string
	expectFailure bool
	success       bool
	connection    string
}

func NewClientTestSuite() *clientTestSuite {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	cts := clientTestSuite{}
	cts.users = []clientCredential{
		{"nopw", "%", "", "mysql_native_password"},
		{"native", "%", "nat", "mysql_native_password"},
		{"sha2", "%", "sha", "caching_sha2_password"},
		{"socket", "%", user.Username, "auth_socket"},
	}
	return &cts
}

func (cts *clientTestSuite) getClients(version string) ([]string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	var clients []string
	if version != "all" {
		clients, err = filepath.Glob(user.HomeDir + "/opt/mysql/" + version + "/bin/mysql")
	} else {
		clients, err = filepath.Glob(user.HomeDir + "/opt/mysql/*/bin/mysql")
	}

	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (cts *clientTestSuite) setupUsers(db *sql.DB) error {
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

func (cts *clientTestSuite) testClient(client string) (clientTestResults, error) {
	var report clientTestResults
	connOpts := [][]string{
		{"-h", "127.0.0.1", "-P", "4000"},
		{"-S", "/tmp/tidb.sock"},
	}
	parts := strings.Split(client, "/")
	clientVer := semver.New(parts[len(parts)-3])

	for _, user := range cts.users {
		for _, connOpt := range connOpts {
			testResult := clientTestResult{
				user:       user,
				client:     client,
				success:    false,
				connection: "TCP",
			}
			if connOpt[0] == "-S" {
				testResult.connection = "socket"
			}

			expectFailure := false

			// auth_socket requires a socket connection to function
			if user.authPlugin == "auth_socket" && connOpt[0] == "-h" {
				expectFailure = true
			}

			// MySQL 5.1.x doesn't support authentication plugins
			if user.authPlugin != "mysql_native_password" && clientVer.LessThan(*semver.New("5.1.99")) {
				expectFailure = true
			}

			// MySQL 5.5 and 5.6 don't ship a caching_sha2_password client plugin
			//
			//     ERROR 2059 (HY000): Authentication plugin 'caching_sha2_password' cannot be loaded:
			//     /usr/local/mysql/lib/plugin/caching_sha2_password.so: cannot open shared object file:
			//     No such file or directory
			if user.authPlugin == "caching_sha2_password" && clientVer.LessThan(*semver.New("5.7.0")) {
				expectFailure = true
			}
			testResult.expectFailure = expectFailure

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
				report.okCount++
				testResult.success = true
			} else if exitCode != 0 && expectFailure {
				if verboseFlag {
					fmt.Printf("\U00002714\U0000FE0F\t")
					fmt.Printf("  Command '%s' returned %d.\n", cmd.String(), cmd.ProcessState.ExitCode())
					println(string(output))
				}
				if user.authPlugin == "caching_sha2_password" {
					if strings.Contains(string(output),
						"ERROR 1251 (08004): Client does not support authentication protocol requested by server; consider upgrading MySQL client") {
					} else if strings.Contains(string(output),
						"ERROR 2059 (HY000): Authentication plugin 'caching_sha2_password' cannot be loaded") {
					} else {
						fmt.Printf("Unexpected output:\n")
						println(string(output))
					}
				}
				report.okCount++
			} else {
				fmt.Printf("\U0000274C\t")
				fmt.Printf("  Command '%s' returned %d.\n", cmd.String(), cmd.ProcessState.ExitCode())
				println(string(output))
				report.failCount++
			}
			report.testResults = append(report.testResults, testResult)
		}
	}

	if verboseFlag {
		fmt.Printf("Testing %s: success=%d, failures=%d\n", client, report.okCount, report.failCount)
	}

	return report, nil
}

func printResults(authMethod string, clients []string, r map[string][]clientTestResult) {
	fmt.Printf("Test results for test with %s as default authentication format\n", authMethod)
	for _, client := range clients[:1] {
		fmt.Printf("%-8s", "auth")
		for _, test := range r[client] {
			switch test.user.authPlugin {
			case "mysql_native_password":
				fmt.Print("\tnative")
			case "caching_sha2_password":
				fmt.Print("\tsha2")
			case "auth_socket":
				fmt.Print("\tsocket")
			}
		}
		fmt.Print("\n")

		fmt.Printf("%-8s", "user")
		for _, test := range r[client] {
			fmt.Printf("\t%s", test.user.user)
		}
		fmt.Print("\n")

		fmt.Printf("%-8s", "connection")
		for _, test := range r[client] {
			fmt.Printf("\t%s", test.connection)
		}
		fmt.Print("\n")
	}
	for _, client := range clients {
		clientName := strings.Split(client, "/")
		fmt.Printf("%-8s", clientName[len(clientName)-3])
		for _, test := range r[client] {
			if test.success && !test.expectFailure {
				fmt.Print("\t\U00002705")
			} else if !test.success && test.expectFailure {
				fmt.Print("\t\U00002714\U0000FE0F")
			} else {
				fmt.Print("\t\U0000274C")
			}
		}
		fmt.Print("\n")

	}
}

func main() {
	flag.BoolVar(&verboseFlag, "verbose", false, "Verbose")
	flag.StringVar(&version, "version", "all", "Version to test")
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

	clients, err := cts.getClients(version)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Clients:\n")
	for _, client := range clients {
		fmt.Printf("\t%s\n", client)
	}

	exitCode := 0
	authMethods := []string{"mysql_native_password", "caching_sha2_password"}
	for _, authMethod := range authMethods {
		clientResults := make(map[string][]clientTestResult)
		fmt.Printf("-----------------------------------------------\n")
		if verboseFlag {
			fmt.Printf("Testing with %s as default authentication plugin\n\n", authMethod)
		}
		_, err = dbAdmin.Exec("SET GLOBAL default_authentication_plugin='" + authMethod + "'")
		if err != nil {
			panic(err)
		}
		for _, client := range clients {
			report, err := cts.testClient(client)
			clientResults[client] = report.testResults
			if err != nil {
				panic(err)
			}
			if report.failCount > 0 {
				exitCode = 1
			}
		}
		printResults(authMethod, clients, clientResults)
	}

	defer dbAdmin.Close()
	os.Exit(exitCode)
}
