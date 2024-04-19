package cmd

import (
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/safe"

	"github.com/stregato/mio/lib/security"
)

var revokeCmd = &assist.Command{
	Use:    "revoke",
	Short:  "Revoke access to a safe",
	Params: []assist.Param{safeParam, existingParam},
	Run: func(args map[string]string) error {
		url := args["safe"]
		userId, _ := security.NewUserId(args["user"])

		s, err := safe.Open(DB, Identity, url)
		if err != nil {
			return err
		}
		defer s.Close()
		groups, err := s.UpdateGroup(safe.UserGroup, safe.Revoke, userId)
		if err != nil {
			return err
		}

		printGroups(groups)

		return nil
	},
}

func init() {
	safeCmd.AddCommand(revokeCmd)
}
