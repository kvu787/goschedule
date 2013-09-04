package goschedule

import (
	"database/sql"
	"fmt"
	"reflect"
)

type Queryer interface {
	TableName() string
}

func Select(db *sql.DB, conditions string, queryer Queryer) ([]Queryer, error) {
	query := fmt.Sprintf("SELECT * FROM %s %s", queryer.TableName(), conditions)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var queryers []Queryer
	for rows.Next() {
		values := []interface{}{}
		rows.Scan(values...)
		// make shadow copy of queryer
		queryer := reflect.New(reflect.TypeOf(queryer)).Elem()
		for i, value := range values {
			queryer.Field(i).Set(reflect.ValueOf(value))
		}
		queryers = append(queryers, queryer.Interface().(Queryer))
	}
	return queryers, nil
}
