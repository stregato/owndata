package cmd

import (
	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/cli/styles"
	"github.com/stregato/stash/lib/core"

	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"
)

var grantCmd = &assist.Command{
	Use:    "grant",
	Short:  "Grant access to a safe",
	Params: []assist.Param{safeParam, userParam},
	Run: func(params map[string]string) error {
		userId, _ := security.CastID(params["user"])

		s, err := getSafeByName(params["safe"])
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
