package cmd

import (
	"github.com/stregato/stash/cli/assist"

	"github.com/stregato/stash/lib/security"
)

var revokeCmd = &assist.Command{
	Use:    "revoke",
	Short:  "Revoke access to a safe",
	Params: []assist.Param{safeParam, existingParam},
	Run: func(params map[string]string) error {
		userId, _ := security.CastID(params["user"])

		s, err := getSafeByName(params["safe"])
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
