package cmd

import "github.com/stregato/mio/cli/assist"

var putCmd = &assist.Command{
	Use:   "put",
	Short: "Put a file into a safe",
	Params: []assist.Param{
		{
			Use:   "safe",
			Short: "The safe to push the file to",
			Complete: func(c *assist.Command, arg string, params map[string]string) {
			},
			Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
				return "", nil
			},
		},
		{
			Use:   "src_file",
			Short: "The file to push",

			Complete: func(c *assist.Command, arg string, params map[string]string) {
			},
			Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
				return "", nil
			},
		},
	},
}

func init() {
	Root.AddCommand(putCmd)
}
