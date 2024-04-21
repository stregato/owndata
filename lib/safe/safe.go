package safe

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
)

type Config struct {
	Quota       int64
	Description string
	Signature   []byte
}

type Safe struct {
	Hnd       int
	ID        string
	URL       string
	DB        *sqlx.DB
	Store     storage.Store
	Config    Config
	CreatorId security.ID
	Identity  *security.Identity
	Lock      sync.RWMutex
}

var DefaultDBPath string
var DefaultDB *sqlx.DB
var DefaultUser *security.Identity

func init() {
	// Get the user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	// Construct the path to the database file
	DefaultDBPath = filepath.Join(configDir, "mio", "mio.db")
}
