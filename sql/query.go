package sql

import (
	s "database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stregato/mio/core"
	"github.com/vmihailenco/msgpack/v5"
)

type Args map[string]any

var ErrNoRows = s.ErrNoRows

func convert(m Args) ([]any, error) {
	var args []any

	for k, v := range m {
		var c any
		switch v := v.(type) {
		case time.Time:
			c = v.UnixNano()
		case string, []byte, int, int8, int16, int32, int64, uint16, uint32, uint64, uint8, float32, float64, bool:
			c = v
		default:
			var err error
			c, err = msgpack.Marshal(v)
			if core.IsErr(err, "cannot marshal attribute %s=%v: %v", k, v, err) {
				return nil, err
			}
		}
		args = append(args, s.Named(k, c))
	}
	return args, nil
}

type Rows struct {
	rows *s.Rows
}

func (db DB) trace(key string, m Args, err error) {
	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		q := db.queries[key]
		for k, v := range m {
			q = strings.ReplaceAll(q, ":"+k, fmt.Sprintf("%v", v))
		}
		logrus.Tracef("SQL: %s: %v", q, err)
	}
}

func (db DB) Exec(key string, m Args) (s.Result, error) {
	args, err := convert(m)
	if err != nil {
		return nil, err
	}
	res, err := db.getStatement(key, m).Exec(args...)
	db.trace(key, m, err)
	if core.IsErr(err, "cannot execute query: %v", err) {
		return nil, err
	}

	return res, nil
}

func (db DB) GetVersion(key string) float32 {
	return db.versions[key]
}

func (db DB) QueryRow(key string, m Args, dest ...any) error {
	args, err := convert(m)
	if err != nil {
		return err
	}
	row := db.getStatement(key, m).QueryRow(args...)
	err = row.Err()
	db.trace(key, m, err)
	if err != s.ErrNoRows && core.IsErr(err, "cannot execute query: %v", err) {
		return err
	}

	return scanRow(row, dest...)
}

func (db DB) Query(key string, m Args) (Rows, error) {
	args, err := convert(m)
	if err != nil {
		return Rows{}, err
	}
	stmt := db.getStatement(key, m)
	rows, err := stmt.Query(args...)
	db.trace(key, m, err)
	return Rows{rows: rows}, err
}

func (db DB) QueryExt(key, sql string, m Args) (Rows, error) {
	args, err := convert(m)
	if err != nil {
		return Rows{}, err
	}
	basic, ok := db.queries[key]
	if !ok {
		return Rows{}, core.Errorf("DbNoKey: missing query %s", key)
	}
	q := basic + " " + sql
	stmt, ok := db.stmts[q]
	if !ok {
		stmt, err = db.Db.Prepare(q)
		if err != nil {
			return Rows{}, core.Errorf("DbPrepare: cannot prepare query %s: %v", q, err)
		}
		db.stmts[q] = stmt
		core.Info("SQL statement compiled: '%s' '%s'\n", key, q)
	}
	rows, err := stmt.Query(args...)
	db.trace(key, m, err)
	return Rows{rows: rows}, err
}

func Map(v any) Args {
	args := Args{}
	val := reflect.ValueOf(v)
	// Handle if the input is a pointer to a struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// Ensure we're dealing with a struct
	if val.Kind() != reflect.Struct {
		return args
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)
		// Use the db tag as the field name if present
		fieldName := structField.Tag.Get("db")
		// If the db tag is "-", exclude the field
		if fieldName == "-" {
			continue
		}
		if fieldName == "" {
			fieldName = structField.Name
		}
		args[fieldName] = field.Interface()
	}

	return args
}

func (rw *Rows) Scan(dest ...interface{}) (err error) {
	columnTypes, err := rw.rows.ColumnTypes()
	if err != nil {
		return err
	}

	for i, col := range columnTypes {
		switch col.DatabaseTypeName() {
		case "INT": // Assuming the column is a Unix timestamp stored as INT
			if t, ok := dest[i].(*time.Time); ok {
				var timestamp int64
				dest[i] = &timestamp
				defer func(index int, originalDest *time.Time) {
					*originalDest = time.Unix(timestamp/1e9, timestamp%1e9)
				}(i, t)
			}

		case "BLOB", "TEXT":
			if _, ok := dest[i].(*string); !ok {
				var data []byte
				var originalDest = dest[i]
				dest[i] = &data
				defer func(index int, originalDest any) {
					if len(data) > 0 {
						err = msgpack.Unmarshal(data, originalDest)
						if err != nil {
							n := reflect.TypeOf(originalDest)
							core.IsErr(err, "cannot convert binary to type %v: %v", n, err)
						}
					}
				}(i, originalDest)
			}
		}
	}

	return rw.rows.Scan(dest...)
}

func scanRow(row *s.Row, dest ...interface{}) (err error) {
	for i, d := range dest {
		switch d := d.(type) {
		case *time.Time:
			var timestamp int64
			var originalDest = d
			dest[i] = &timestamp
			defer func(index int, originalDest *time.Time) {
				*originalDest = time.Unix(timestamp/1e9, timestamp%1e9)
			}(i, originalDest)
		case *string, *[]byte, *int, *int8, *int16, *int32, *int64, *uint16, *uint32, *uint64, *uint8, *float32, *float64, *bool:
			continue
		default:
			var data []byte
			var originalDest = dest[i]
			dest[i] = &data
			defer func(index int, originalDest any) {
				if len(data) > 0 {
					err = msgpack.Unmarshal(data, originalDest)
					if err != nil {
						n := reflect.TypeOf(originalDest)
						core.IsErr(err, "cannot convert binary to type %v: %v", n, err)
					}
				}
			}(i, originalDest)
		}
	}

	return row.Scan(dest...)
}

func (rw *Rows) Next() bool {
	return rw.rows.Next()
}

func (rw *Rows) Close() error {
	return rw.rows.Close()
}
