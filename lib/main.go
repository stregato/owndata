package main

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/sqlx"
)

func main() {
	fmt.Print("This is just a library! ")
}

var db *sqlx.DB
var safes map[int]*safe.Safe
var safesCnt int
var safesCntLock sync.RWMutex

func OpenSafe(url, creator string) (*safe.Safe, error) {
	return nil, nil
}
