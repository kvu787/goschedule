package main

import (
	"database/sql"
	"fmt"

	"github.com/kvu787/go_schedule/backend/config"
	"github.com/kvu787/go_schedule/backend/database"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open(config.Db, config.DbConn)
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	toAdd := database.Dept{"i", "like", "pie"}
	if err = database.Insert(db, toAdd); err != nil {
		fmt.Println(err)
	}
	res, err := database.Select(db, database.Dept{}, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range res {
		fmt.Println(v)
	}
}
