// Package database provides basic (read: fragile)
// object relational mapping between structs and
// database tables.
package database

import (
	"database/sql"
	"fmt"
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
		fmt.Println("Failed to begin transaction")
		tx.Rollback()
		return nil, err
	}
	query := fmt.Sprintf("SELECT * FROM %s %s", q.TableName(), filters)
	rows, err := tx.Query(query)
	if err != nil {
		fmt.Println("Query failed")
		tx.Rollback()
		return nil, err
	}
	qs, err := receive(q, rows)
	if err != nil {
		fmt.Println("Error in storing queried rows into a struct")
		fmt.Println(err)
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
		fmt.Println("Failed to begin transaction")
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

// Update uses the PrimaryKey and TableName methods
// of q to update its fields in the database.
// Form:
// UPDATE [q.TableName] set ... WHERE [PrimaryKeyField] = [PrimaryKey]
func Update(db *sql.DB, q Queryer) error {
	tx, err := db.Begin()
	defer tx.Commit()
	if err != nil {
		fmt.Println("Failed to begin transaction")
		tx.Rollback()
		return err
	}
	sql := prepareInsertString(q)
	if _, err := tx.Exec(sql, prepareInsertArguments(q)...); err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to update records: %s", err)
	}
	tx.Commit()
	return nil
}
