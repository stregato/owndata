package cmd

import "github.com/stregato/stash/cli/assist"

var chatCmd = &assist.Command{
	Use:   "chat",
	Short: "Send and receive messages",
}

func init() {
	Root.AddCommand(chatCmd)
}
