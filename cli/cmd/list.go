package cmd

import (
	"fmt"

	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/cli/styles"
	"github.com/stregato/stash/lib/security"
)

type safeDesc struct {
	name      string
	creator   string
	creatorId security.ID
	url       string
}

var listCmd = &assist.Command{
	Use:   "list",
	Short: "List all safes",
	Run: func(params map[string]string) error {

		descs, err := listSafes()
		if err != nil {
			return err
		}
		for _, desc := range descs {
			fmt.Println(styles.UseStyle.Render(desc.name), " by ", styles.ShortStyle.Render(desc.creator))
			fmt.Println("  ", desc.url)
			fmt.Println()
		}
		return nil
	},
}

func init() {
	safeCmd.AddCommand(listCmd)
}
