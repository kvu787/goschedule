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

// Select uses an empty struct to query the database and return a slice of the
// corresponding structs.
// It returns records from a SQL query in the form:
//
//	'SELECT * FROM <lowercase struct name> [additional SQL clauses]...'
//
// Additional SQL clauses can be specified in the filters parameter.
// For example `select(db, Sect{}, "ORDER BY sln LIMIT 5")` runs the query:
//
//	SELECT * FROM sect ORDER BY sln LIMIT 5;
func Select(db *sql.DB, object interface{}, conditions string) ([]interface{}, error) {
	tableName := reflect.TypeOf(object).Name()
	query := fmt.Sprintf("SELECT * FROM %s %s", tableName, conditions)
	// execute query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	// get column names (undercase)
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var records []interface{}
	// store rows in records
	for rows.Next() {
		objectValue := reflect.ValueOf(object)
		values := []interface{}{}
		// prepare values slice to receive row values
		for _, columnName := range columnNames {
			column := objectValue.FieldByNameFunc(func(name string) bool {
				return strings.ToLower(name) == strings.ToLower(columnName)
			}).Interface()
			values = append(values, &column)
		}
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		// scan row values into slice
		rows.Scan(values...)
		// create a new object and assign values to fields according to column names
		object := reflect.New(reflect.TypeOf(object)).Elem()
		for i, value := range values {
			field := object.FieldByNameFunc(func(name string) bool {
				return strings.ToLower(name) == strings.ToLower(columnNames[i])
			})
			valueInterface := reflect.ValueOf(value).Elem().Interface()
			switch valueInterface.(type) {
			case []uint8:
				field.SetString(string(valueInterface.([]uint8)))
			case int64:
				field.SetInt(valueInterface.(int64))
			default:
				panic(fmt.Sprintf("goschedule.Select error: type unsupported `%T`. Edit goschedule.Select source to implement the type or use a supported type.", valueInterface))
			}
		}
		records = append(records, object.Interface())
	}
	return records, nil
}

// Insert inserts a struct into the database. It ignores struct fields with the tag `ignore:"true"`.
// It returns records from a SQL query in the form:
//
//	'INSERT INTO <lowercase struct name> VALUES (<non-ignored struct field values, in sequential order)'
//
// For example, the following:
//
//	type Banana struct {
//		Color     string
// 		Length    int
// 		index     int
// 		Throwable bool `ignore:"true"`
//	}
//
// 	Insert(db, Banana{"yellow", 12, 0, true})
//
// will run the query:
//
// 	INSERT INTO banana VALUES ('yellow', 12, true);
//
// Note that unexported fields are still added to the query unless they have the ignore tag.
func Insert(db *sql.DB, object interface{}) error {
	// prepare columns names
	objectType := reflect.TypeOf(object)
	objectValue := reflect.ValueOf(object)
	// prepare values and placeholder string
	var values []interface{}
	var placeholder []string
	for i := 0; i < objectType.NumField(); i++ {
		if objectType.Field(i).Tag.Get("ignore") != "true" {
			values = append(values, objectValue.Field(i).Interface())
		}
	}
	for i, _ := range values {
		placeholder = append(placeholder, fmt.Sprintf("$%d", i+1))
	}
	// execute query
	query := fmt.Sprintf("INSERT INTO %s VALUES (%s)", objectType.Name(), strings.Join(placeholder, ","))
	if _, err := db.Exec(query, values...); err != nil {
		return fmt.Errorf("Failed to insert records: %s", err)
	}
	return nil
}
