package sqlx

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stregato/mio/lib/core"
)

func (db *DB) Define(version float32, ddl string) error {
	parts := strings.Split(ddl, "\n")

	var header string
	for i := 0; i < len(parts); i++ {
		part := strings.Trim(parts[i], " ")
		if len(part) == 0 {
			continue
		}
		if strings.HasPrefix(part, "-- ") {
			header = strings.Trim(part[3:], " ")
		} else {
			var query string
			line := i
			for ; i < len(parts); i++ {
				part := strings.Trim(parts[i], " ")
				if len(part) == 0 {
					break
				}
				query += part + "\n"
			}
			if header == "INIT" {
				_, err := db.Db.Exec(query)
				if core.IsErr(err, "cannot execute SQL Init stmt (line %d) '%s': %v\n", line, query, err) {
					return err
				}
				core.Info("SQL Init stmt (line %d) '%s' executed\n", line, query)
			} else {
				err := db.prepareStatement(version, header, query, line)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (db *DB) Keys() []string {
	var keys []string

	for k := range db.stmts {
		keys = append(keys, k)
	}
	return keys
}

func (db *DB) prepareStatement(version float32, key, s string, line int) error {
	key = strings.Trim(key, " ")
	v, ok := db.versions[key]
	if ok {
		switch {
		case v == version:
			if db.queries[key] != s {
				logrus.Panicf("duplicate SQL statement for key '%s' (line %d)\n", s, line)
				panic(key)
			}
		case v < version:
			return nil
		}
	}

	if !strings.Contains(s, "#") {
		stmt, err := db.Db.Prepare(s)
		if core.IsErr(err, "cannot compile SQL statement '%s' (%d) '%s': %v\n", key, line, s) {
			return err
		}
		db.stmts[key] = stmt
		core.Info("SQL statement compiled: '%s' (%d) '%s'\n", key, line, s)
	}

	db.queries[key] = s
	db.versions[key] = version
	return nil
}

func (db *DB) getStatement(key string, args Args) *sql.Stmt {
	if v, ok := db.stmts[key]; ok {
		return v
	}

	if s, ok := db.queries[key]; ok {
		s = replaceArgs(s, args)
		if v, ok := db.stmts[s]; ok {
			return v
		}

		stmt, err := db.Db.Prepare(s)
		if err != nil {
			logrus.Panicf("cannot compile SQL statement for key '%s': %v", key, err)
			panic(key)
		}
		db.stmts[s] = stmt
		return stmt
	}

	logrus.Panicf("unknown SQL statement for key '%s'", key)
	panic(key)
}

func replaceArgs(s string, args Args) string {

	// Compile a regular expression that matches words starting with '#'
	re := regexp.MustCompile(`#\w+`)

	// Use the ReplaceAllStringFunc method to replace matches using a custom function
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		// Look up the key in the map. If found, return its value; otherwise, return the match unchanged.
		if val, ok := args[match]; ok {
			if ss, ok := val.(string); ok {
				return ss
			}
		}
		return match
	})

	return result
}
