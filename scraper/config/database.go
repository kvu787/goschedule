package config

import (
	"fmt"
)

// A DbConn provides data to create a *sql.DB connection for
// PostgreSQL databases.
type DbConn struct {
	driver   string
	user     string
	dbname   string
	password string
	sslmode  string
}

// Driver returns the driver attribute of DbConn.
func (db DbConn) Driver() string {
	return db.driver
}

// Conn returns a connection string using the appropriate
// attributes of the DBConn.
func (db DbConn) Conn() string {
	return fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=%s",
		db.user,
		db.dbname,
		db.password,
		db.sslmode,
	)
}

var (
	// DbConnSwitch provides a connection string to the 'switch db',
	// which indicates which database should be scraped and used
	// for web requests.
	// The web app should use whichever db that the scraper is not
	// using at any given time.
	DbConnSwitch DbConn = DbConn{
		"postgres",
		"gosh",
		"switch_db",
		"gosh",
		"require",
	}

	// DbConn1 provides a connection string to one of two
	// application databases, which store all schedule information.
	// If DbConn1 is being used by the scraper, DbConn2 should be
	// used to serve web requests, and vice versa.
	DbConn1 DbConn = DbConn{
		"postgres",
		"gosh",
		"gosh1",
		"gosh",
		"require",
	}

	// The other application database. See documentation for DBConn1.
	DbConn2 DbConn = DbConn{
		"postgres",
		"gosh",
		"gosh2",
		"gosh",
		"require",
	}

	// TestDbConn is by Go tests.
	TestDbConn DbConn = DbConn{
		"postgres",
		"gosh_test",
		"gosh",
		"gosh",
		"require",
	}
)
