package cmd

import (
	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/fs"
)

func getRun(args map[string]string) error {
	s, err := getSafeByName(args["safe"])
	if err != nil {
		return err
	}

	f, err := fs.Open(s)
	if err != nil {
		return err
	}

	_, err = f.GetFile(args["src"], args["dst"], fs.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

var getCmd = &assist.Command{
	Use:   "get",
	Short: "Get a file from the safe",
	Params: []assist.Param{
		safeParam,
		srcParam,
		destParam,
	},
	Run: getRun,
}

func init() {
	filesCmd.AddCommand(getCmd)
}
