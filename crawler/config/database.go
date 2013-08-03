package config

import (
	"fmt"
)

// These constants are used to create the database
// connection strings.
// They should be changed by each user.
const (
	Db       string = "postgres"
	dbname   string = "gosh"
	user     string = "gosh"
	password string = "gosh"
	sslmode  string = "require"

	TestDb       string = "postgres"
	testdbname   string = "gosh_test"
	testuser     string = "gosh"
	testpassword string = "gosh"
	testsslmode  string = "require"
)

// DbConn and TestDbConn provide database connection
// strings based on the constants in this file.
var (
	DbConn string = fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=%s",
		user,
		dbname,
		password,
		sslmode,
	)
	TestDbConn string = fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=%s",
		testuser,
		testdbname,
		testpassword,
		testsslmode,
	)
)
