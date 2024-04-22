package cmd

import "github.com/stregato/mio/cli/assist"

var safeCmd = &assist.Command{
	Use:   "config",
	Short: "Manage safes and users",
}

func init() {
	Root.AddCommand(safeCmd)
}
