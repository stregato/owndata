package cmd

import (
	"strconv"

	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/cli/styles"
	"github.com/stregato/mio/lib/fs"
)

func lsRun(params map[string]string) error {
	s, dir, err := getSafeAndPath(params["dir"])
	if err != nil {
		return err
	}
	defer s.Close()

	f, err := fs.Open(s)
	if err != nil {
		return err
	}
	defer f.Close()

	files, err := f.List(dir, fs.ListOptions{})
	if err != nil {
		return err
	}

	println(styles.ShortStyle.Render(params["dir"]))
	for _, file := range files {
		name := file.Name
		if name == "" {
			name = "."
		}
		creator := file.Creator.Nick()
		if creator == "" {
			creator = "-"
		}
		println(styles.UseStyle.Render(name), styles.ShortStyle.Render(strconv.Itoa(file.Size)),
			styles.ShortStyle.Render(creator), styles.ShortStyle.Render(file.ModTime.String()),
			styles.ShortStyle.Render((file.LocalCopy)))
	}

	return nil
}

var lsDirParam = assist.Param{
	Use:   "dir",
	Short: "The directory to list",
	Complete: pathComplete(pathMatchOptions{
		safePath: true,
	}),
	Match: pathMatch(pathMatchOptions{
		msg:      "Select a directory",
		safePath: true,
	}),
}

var lsCmd = &assist.Command{
	Use:   "ls",
	Short: "List files in the folder",
	Params: []assist.Param{
		lsDirParam,
	},
	Run: lsRun,
}

func init() {
	filesCmd.AddCommand(lsCmd)
}
