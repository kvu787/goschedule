package database

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

// Receive stores the results of a raw SQL select
// into the appropriate struct that implements
// Queryer.
// How/why the hell does this work
func receive(q Queryer, rows *sql.Rows) ([]Queryer, error) {
	var queryers []Queryer
	for rows.Next() {
		qNew := q
		value := reflect.ValueOf(q)
		baseType := reflect.TypeOf(value.Interface())
		tmp := reflect.New(baseType)
		ptrValue := tmp.Elem()
		res := []interface{}{}
		for i := 0; i < value.NumField(); i++ {
			tmp := value.Field(i).Interface()
			res = append(res, &tmp)
		}
		if err := rows.Scan(res...); err != nil {
			return nil, err
		}
		for i := 0; i < value.NumField(); i++ {
			underlyingValue := reflect.ValueOf(res[i]).Elem()
			switch v := underlyingValue.Interface().(type) {
			case string:
				ptrValue.Field(i).SetString(string(v))
				break
			case []byte:
				ptrValue.Field(i).SetString(string(v))
				break
			case int64:
				ptrValue.Field(i).SetInt(int64(v))
				break
			case nil:
				break
			default:
				return nil, errors.New(
					fmt.Sprintf("Failed to fetch value from database: %v", underlyingValue.Interface()),
				)
			}
		}
		qNew = ptrValue.Interface().(Queryer)
		queryers = append(queryers, qNew)
	}
	return queryers, nil
}

// TODO (kvu787): how are nil values handled?
func prepareInsertString(q Queryer) string {
	sql := fmt.Sprintf("INSERT INTO %s VALUES (", q.TableName())
	v := reflect.ValueOf(q)
	for i := 0; i < v.NumField(); i++ {
		if i == v.NumField()-1 {
			sql += fmt.Sprintf("$%d)", i+1)
			break
		}
		sql += fmt.Sprintf("$%d,", i+1)
	}
	return sql
}

func prepareInsertArguments(q Queryer) []interface{} {
	v := reflect.ValueOf(q)
	arguments := []interface{}{}
	for i := 0; i < v.NumField(); i++ {
		arguments = append(arguments, v.Field(i).Interface())
	}
	return arguments
}
