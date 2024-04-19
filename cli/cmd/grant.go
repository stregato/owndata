package cmd

import (
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/cli/styles"
	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"

	"github.com/stregato/mio/lib/security"
)

var grantCmd = &assist.Command{
	Use:    "grant",
	Short:  "Grant access to a safe",
	Params: []assist.Param{safeParam, userParam},
	Run: func(args map[string]string) error {
		url := args["safe"]
		userId, _ := security.NewUserId(args["user"])

		s, err := safe.Open(DB, Identity, url)
		if err != nil {
			return err
		}
		defer s.Close()
		groups, err := s.UpdateGroup(safe.UserGroup, safe.Grant, userId)
		if err != nil {
			return err
		}

		token, err := security.EcEncrypt(userId, []byte(s.URL))
		if err != nil {
			return err
		}

		println(styles.UseStyle.Render("Token"), styles.ShortStyle.Render(core.EncodeBinary(token)))
		println()
		printGroups(groups)

		return nil
	},
}

func init() {
	safeCmd.AddCommand(grantCmd)
}
