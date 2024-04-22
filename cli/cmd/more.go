package cmd

import (
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/fs"
)

func moreRun(args map[string]string) error {
	s, err := getSafeByNameOrUrl(args["safe"])
	if err != nil {
		return err
	}

	f, err := fs.Open(s)
	if err != nil {
		return err
	}

	data, err := f.GetData(args["src"], fs.GetOptions{})
	if err != nil {
		return err
	}
	println(string(data))
	return nil
}

var moreCmd = &assist.Command{
	Use:   "more",
	Short: "Show the content of a file in the safe",
	Params: []assist.Param{
		safeParam,
		srcParam,
	},
	Run: moreRun,
}

func init() {
	filesCmd.AddCommand(moreCmd)
}
