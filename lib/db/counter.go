package db

import (
	"strings"

	"github.com/stregato/stash/lib/sqlx"
)

func (d *DB) GetCounter(table string, key string) (int, error) {
	table = strings.ToUpper(table)

	err := d.createCounterTableIfNeeded(table)
	if err != nil {
		return 0, err
	}
	rows, err := d.DB.Query("SELECT SUM(VALUE) FROM "+table+" WHERE KEY = :key", sqlx.Args{"key": key})
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var counter int
	for rows.Next() {
		err = rows.Scan(&counter)
		if err != nil {
			return 0, err
		}
	}
	return counter, nil
}

func (t *Transaction) IncCounter(table string, key string, value int) error {
	table = strings.ToUpper(table)

	err := t.db.createCounterTableIfNeeded(table)
	if err != nil {
		return err
	}

	ql := "INSERT INTO " + table + " (KEY, VALUE) VALUES (:key, :value)"
	_, err = t.Exec(ql, sqlx.Args{"key": key, "value": value})
	if err != nil {
		return err
	}
	return err
}

func (d *DB) createCounterTableIfNeeded(table string) error {
	if d.counters[table] {
		return nil
	}

	_, err := d.DB.Db.Exec("CREATE TABLE IF NOT EXISTS " + table + " (KEY TEXT, VALUE INTEGER)")
	if err != nil {
		return err
	}
	_, err = d.DB.Db.Exec("CREATE INDEX IF NOT EXISTS " + table + "_key ON " + table + " (KEY)")
	if err != nil {
		return err
	}

	d.counters[table] = true
	return err
}
