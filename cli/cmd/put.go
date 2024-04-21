package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/fs"
)

var srcParam = assist.Param{
	Use:   "src",
	Short: "The source path",

	Complete: func(c *assist.Command, arg string, params map[string]string) {
	},
	Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
		if arg == "" {
			survey.AskOne(&survey.Input{
				Message: "Enter the path of the source file",
			}, &arg)
		}
		return arg, nil
	},
}

var destParam = assist.Param{
	Use:   "dst",
	Short: "The destination path",
	Complete: func(c *assist.Command, arg string, params map[string]string) {
	},
	Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
		if arg == "" {
			survey.AskOne(&survey.Input{
				Message: "Enter the destination path",
			}, &arg)
		}
		return arg, nil
	},
}

func putRun(args map[string]string) error {
	s, err := getSafeByNameOrUrl(args["safe"])
	if err != nil {
		return err
	}

	f, err := fs.Open(s)
	if err != nil {
		return err
	}

	_, err = f.PutFile(args["dst"], args["src"], fs.PutOptions{})
	if err != nil {
		return err
	}
	return nil
}

var putCmd = &assist.Command{
	Use:   "put",
	Short: "Put a file into a safe",
	Params: []assist.Param{
		safeParam,
		srcParam,
		destParam,
	},
	Run: putRun,
}

func init() {
	Root.AddCommand(putCmd)
}
