package sql

import (
	_ "embed"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stregato/mio/core"
)

//go:embed test.sql
var testDDL string

type content struct {
	Title string
	Text  string
}

func TestDb(t *testing.T) {
	db, err := Open(":memory:")
	core.TestErr(t, err, "cannot open memory db: %v", err)

	err = db.Define(1.0, testDDL)
	core.TestErr(t, err, "cannot define db: %v", err)

	keys := db.Keys()
	core.Assert(t, len(keys) == 4, "keys is not 4 size")

	now := time.Now()
	_, err = db.Exec("INSERT_USER", Args{"username": "john", "email": "Email", "registrationDate": now})
	core.TestErr(t, err, "cannot insert user: %v", err)

	_, err = db.Exec("INSERT_POST", Args{"userId": 1, "title": "Title", "content": content{"title", "text"},
		"postDate": now})
	core.TestErr(t, err, "cannot insert post: %v", err)

	rows, err := db.Query("SELECT_POSTS", Args{"username": "john"})
	core.TestErr(t, err, "cannot select posts: %v", err)

	var (
		title    string
		cnt      content
		postDate time.Time
	)
	rows.Next()
	err = rows.Scan(&title, &cnt, &postDate)
	core.TestErr(t, err, "cannot scan posts: %v", err)
	core.Assert(t, postDate.Equal(now), "post date is not equal")
	core.Assert(t, cnt.Title == "title", "content title is not title")
	core.Assert(t, cnt.Text == "text", "content text is not text")

	err = db.Close()
	core.TestErr(t, err, "cannot close db: %v", err)

}
