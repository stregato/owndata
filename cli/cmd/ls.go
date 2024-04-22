package cmd

import (
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/fs"
)

var dirParam = assist.Param{
	Use:   "dir",
	Short: "The directory to list",
	Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
		return arg, nil
	},
}

func lsRun(params map[string]string) error {
	s, err := getSafeByNameOrUrl(params["safe"])
	if err != nil {
		return err
	}
	defer s.Close()

	dir := params["dir"]

	f, err := fs.Open(s)
	if err != nil {
		return err
	}
	defer f.Close()

	files, err := f.List(dir, fs.ListOptions{})
	if err != nil {
		return err
	}

	for _, file := range files {
		println(file.Name)
	}

	return nil
}

var lsCmd = &assist.Command{
	Use:   "ls",
	Short: "List files in the folder",
	Params: []assist.Param{
		safeParam,
		dirParam,
	},
	Run: lsRun,
}

func init() {
	filesCmd.AddCommand(lsCmd)
}
