package goschedule

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func GenerateSchema(s interface{}) string {
	sType := reflect.TypeOf(s)
	sValue := reflect.ValueOf(s)
	if sType.Name() == "" {
		panic("goschedule.GenerateSchema: cannot generate schema for anonymous struct")
	}
	columns := ""
	for i := 0; i < sType.NumField(); i++ {
		// check if ignored
		switch ignore := sType.Field(i).Tag.Get("ignore"); ignore {
		case "true":
			continue
		case "":
		default:
			panic(fmt.Sprintf("goschedule.GenerateSchema: invalid value for 'ignore' key in struct field tag: %q", ignore))
		}
		// add column name
		columns += sType.Field(i).Name
		// add column type
		switch field := sValue.Field(i).Interface(); field.(type) {
		case string:
			columns += " text"
		case int64:
			columns += " integer"
		default:
			panic(fmt.Sprintf("goschedule.GenerateSchema: invalid struct field type: %T", field))
		}
		// add PRIMARY KEY restraint if found
		switch pk := sType.Field(i).Tag.Get("pk"); pk {
		case "true":
			columns += " PRIMARY KEY"
		case "":
		default:
			panic(fmt.Sprintf("goschedule.GenerateSchema: invalid value for 'pk' key in struct field tag: %q", pk))
		}
		// add REFERENCES (foreign key) restraint if found
		if fk := sType.Field(i).Tag.Get("fk"); fk != "" {
			columns += " REFERENCES " + fk
		}
		// add comma
		columns += ", "
	}
	// strip trailing comma
	columns = strings.TrimSuffix(strings.TrimSpace(columns), ",")
	return fmt.Sprintf(
		"CREATE TABLE %s (%s);",
		sType.Name(),
		columns,
	)
}

func Select(db *sql.DB, object interface{}, conditions string) ([]interface{}, error) {
	tableName := reflect.TypeOf(object).Name()
	query := fmt.Sprintf("SELECT * FROM %s %s", tableName, conditions)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var records []interface{}
	for rows.Next() {
		values := []interface{}{}
		rows.Scan(values...)
		object := reflect.New(reflect.TypeOf(object)).Elem()
		for i, value := range values {
			object.Field(i).Set(reflect.ValueOf(value))
		}
		records = append(records, object.Interface())
	}
	return records, nil
}

func Insert(db *sql.DB, object interface{}) error {
	// prepare columns names
	oType := reflect.TypeOf(object)
	oValue := reflect.ValueOf(object)
	// prepare values and placeholder string
	var values []interface{}
	var placeholder []string
	for i := 0; i < oType.NumField(); i++ {
		if oType.Field(i).Tag.Get("ignore") != "true" {
			values = append(values, oValue.Field(i).Interface())
		}
	}
	for i, _ := range values {
		placeholder = append(placeholder, fmt.Sprintf("$%d", i+1))
	}
	// execute query
	query := fmt.Sprintf("INSERT INTO %s VALUES (%s)", oType.Name(), strings.Join(placeholder, ","))
	if _, err := db.Exec(query, values...); err != nil {
		return fmt.Errorf("Failed to insert records: %s", err)
	}
	return nil
}
