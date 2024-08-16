package cmd

import (
	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/fs"
)

func moreRun(args map[string]string) error {
	s, src, err := getSafeAndPath(args["src"])
	if err != nil {
		return err
	}
	defer s.Close()

	f, err := fs.Open(s)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := f.GetData(src, fs.GetOptions{})
	if err != nil {
		return err
	}
	println(string(data))
	return nil
}

var morePathParam = assist.Param{
	Use:   "src",
	Short: "The file to show",
	//Match: safeFileMatch,
}

var moreCmd = &assist.Command{
	Use:   "more",
	Short: "Show the content of a file in the safe",
	Params: []assist.Param{
		morePathParam,
	},
	Run: moreRun,
}

func init() {
	filesCmd.AddCommand(moreCmd)
}
