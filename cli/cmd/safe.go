package cmd

import "github.com/stregato/mio/cli/assist"

var safeCmd = &assist.Command{
	Use:   "safe",
	Short: "Manage safes",
}

func init() {
	Root.AddCommand(safeCmd)
}
