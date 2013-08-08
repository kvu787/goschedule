// Package database provides basic (read: fragile)
// object relational mapping between structs and
// database tables.
package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"

	"github.com/kvu787/go-schedule/scraper/config"
	_ "github.com/lib/pq"
)

// Select uses an empty Queryer struct to query the database
// and return a slice of the corresponding Queryer structs.
// By default it returns records in the form:
// 'SELECT * FROM [tableName]'.
// Additional SQL clauses can be specified in the filters
// parameter.
// TODO (kvu787): might want to split this into seperate functions
func Select(db *sql.DB, q Queryer, filters string) ([]Queryer, error) {
	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	query := fmt.Sprintf("SELECT * FROM %s %s", q.TableName(), filters)
	rows, err := tx.Query(query)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	qs, err := receive(q, rows)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return qs, nil
}

// Insert inserts a Queryer struct into the database.
func Insert(db *sql.DB, q Queryer) error {
	tx, err := db.Begin()
	defer tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	sql := prepareInsertString(q)
	if _, err := tx.Exec(sql, prepareInsertArguments(q)...); err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to insert records: %s", err)
	}
	tx.Commit()
	return nil
}

func ParseSqlFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	var statements []string
	var statement string
	for scanner.Scan() {
		l := scanner.Text()
		for i, w := range l {
			if w == '-' && l[i+1] == '-' {
				break
			} else if w == ';' {
				statement += string(w)
				statements = append(statements, statement)
				statement = ""
			} else {
				statement += string(w)
			}
		}
	}
	return statements, nil
}

func GetSwitch(db *sql.DB) (int, error) {
	var result int
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 1", config.SwitchDBCol, config.SwitchDBTable)
	if err := db.QueryRow(query).Scan(&result); err != nil {
		return -1, err
	}
	return result, nil
}

func FlipSwitch(db *sql.DB) error {
	currentSwitch, err := GetSwitch(db)
	if err != nil {
		return err
	}
	var newSwitch int
	if currentSwitch == 0 {
		newSwitch = 1
	} else {
		newSwitch = 0
	}
	query := fmt.Sprintf(
		"UPDATE %s SET %s = %d WHERE %s = %d",
		config.SwitchDBTable,
		config.SwitchDBCol,
		newSwitch,
		config.SwitchDBCol,
		currentSwitch,
	)
	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
