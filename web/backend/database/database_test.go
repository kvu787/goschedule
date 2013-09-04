package database

import (
	"database/sql"
	"testing"

	"github.com/kvu787/go-schedule/scraper/config"
	_ "github.com/lib/pq"
)

func TestTestDbConnection(t *testing.T) {
	db, err := sql.Open(config.TestDb, config.TestDbConn)
	if err != nil {
		t.Log(err)
		t.Fatalf("Test db config string is improperly formatted, stopping test execution")
	}
	err = db.Ping()
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not connect to test db with config, stopping test execution")
	}
}

// func TestStore(t *testing.T) {

// }

// func TestGet(t *testing.T) {
// 	db, _ := sql.Open(config.TestDb, config.TestDbConn)
// 	setupDb(db)
// 	defer cleanupDb(db)

// }

// func TestInsertSQL(t *testing.T) {
// 	dept := Dept{"School of Hard Knocks", "SHK", "shk.com"}
// 	actual := dept.InsertSQL()
// 	expected := "INSERT INTO depts VALUES ('School of Hard Knocks', 'SHK', 'shk.com')"
// 	if actual != expected {
// 		t.Errorf("Expected: %s, Actual: %s\n", expected, actual)
// 	}
// }

// func setupDb(db *sql.DB) {
// 	db.Exec(`
// 		CREATE TABLE sects (
// 			sln text,
// 			instructor text,
// 			open_spots integer
// 		);
// 	`)
// }

// func cleanupDb(db *sql.DB) {
// 	db.Exec(`
// 		DROP TABLE sects;
// 	`)
// }

// var sect testSect = testSect{"123132", "George Wu", 12}

// type testSect struct {
// 	sln        string
// 	instructor string
// 	openSpots  int
// }

// type testSects []testSect
