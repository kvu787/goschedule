package shared

import (
	"database/sql"
	"fmt"
)

// GetSwitch queries the 'switch db' returns either 1 or 2.
// Used to determine which database should be used to store scrape results.
func GetSwitch(db *sql.DB) (int, error) {
	var result int
	query := fmt.Sprintf("SELECT switch_col FROM switch_table LIMIT 1")
	if err := db.QueryRow(query).Scan(&result); err != nil {
		return -1, err
	}
	return result, nil
}
