package config

import (
	"fmt"
)

type DbConn struct {
	driver   string
	user     string
	dbname   string
	password string
	sslmode  string
}

func (db DbConn) Driver() string {
	return db.driver
}

func (db DbConn) Conn() string {
	return fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=%s",
		db.user,
		db.dbname,
		db.password,
		db.sslmode,
	)
}

const (
	SchemaPath string = "utility/sql/schema.sql" // relative to crawler.go
)

var (
	Switch DbConn = DbConn{
		"postgres",
		"gosh",
		"switch_db",
		"gosh",
		"require",
	}

	DbConn1 DbConn = DbConn{
		"postgres",
		"gosh",
		"gosh1",
		"gosh",
		"require",
	}
	DbConn2 DbConn = DbConn{
		"postgres",
		"gosh",
		"gosh2",
		"gosh",
		"require",
	}
	TestDbConn DbConn = DbConn{
		"postgres",
		"gosh_test",
		"gosh",
		"gosh",
		"require",
	}
)
