package safe

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
)

type Safe struct {
	Id          int
	Db          *sqlx.DB
	Store       storage.Store
	CreatorId   security.UserId
	CurrentUser *security.Identity
	Lock        sync.RWMutex
}

var defaultDB *sqlx.DB
var defaultDBPath string
var currentUser *security.Identity

func init() {
	// Get the user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	// Construct the path to the database file
	defaultDBPath = filepath.Join(configDir, "mio", "mio.db")
}

func CopySafe(c *Safe) *Safe {
	return &Safe{
		Db:          c.Db,
		Store:       c.Store,
		CreatorId:   c.CreatorId,
		CurrentUser: c.CurrentUser,
	}
}
