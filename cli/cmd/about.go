package cmd

import (
	"fmt"

	"github.com/stregato/mio/cli/assist"
)

var aboutId = &assist.Command{
	Use:   "id",
	Short: "show your user public id",
	Run: func(args map[string]string) error {
		fmt.Println(Identity.Id)
		return nil
	},
}

var aboutNick = &assist.Command{
	Use:   "nick",
	Short: "show your user nickname",
	Run: func(args map[string]string) error {
		fmt.Println(Identity.Id.Nick())
		return nil
	},
}

var aboutPrivate = &assist.Command{
	Use:   "private",
	Short: "show your private key",
	Run: func(args map[string]string) error {
		fmt.Println(Identity.Private)
		return nil
	},
}

var aboutDb = &assist.Command{
	Use:   "db",
	Short: "show the path to the database",
	Run: func(args map[string]string) error {
		fmt.Println(DBPath)
		return nil
	},
}

var about = &assist.Command{
	Use:   "about",
	Short: "Show information about current setup",

	Subcommands: []*assist.Command{aboutId, aboutNick, aboutPrivate, aboutDb},
}

func init() {
	Root.AddCommand(about)
}
