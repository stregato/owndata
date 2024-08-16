package cmd

import "github.com/stregato/stash/cli/assist"

var filesCmd = &assist.Command{
	Use:   "files",
	Short: "Upload, download, and manage files",
}

func init() {
	Root.AddCommand(filesCmd)
}
