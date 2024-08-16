package cmd

import (
	"path/filepath"

	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/fs"
)

var srcParam = assist.Param{
	Use:   "src",
	Short: "The source path on the local filesystem",

	Complete: pathComplete(pathMatchOptions{
		onlyFile: true,
	}),
	Match: pathMatch(pathMatchOptions{
		msg: "Select a source",
	}),
}

var destParam = assist.Param{
	Use:   "dst",
	Short: "The destination path on the safe",
	Complete: pathComplete(pathMatchOptions{
		safePath: true,
	}),
	Match: pathMatch(pathMatchOptions{
		msg:      "Select a destination",
		safePath: true,
	}),
}

func putRun(args map[string]string) error {
	s, dst, err := getSafeAndPath(args["dst"])
	if err != nil {
		return err
	}
	defer s.Close()
	if dst == "" {
		dst = filepath.Base(args["src"])
	}

	f, err := fs.Open(s)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.PutFile(dst, args["src"], fs.PutOptions{})
	if err != nil {
		return err
	}
	return nil
}

var putCmd = &assist.Command{
	Use:   "put",
	Short: "Put a file into a safe",
	Params: []assist.Param{
		srcParam,
		destParam,
	},
	Run: putRun,
}

func init() {
	filesCmd.AddCommand(putCmd)
}
