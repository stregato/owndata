package commands

import (
	"github.com/stregato/mio/cli/assist"
)

var RootCmd = &assist.Command{
	Use:   "mio",
	Short: "mio is a CLI tool to manage encrypted safes on remote servers",
}
