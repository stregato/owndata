package cmd

import (
	"net/url"
	"path"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/cli/styles"
	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/stash"
)

var askUrl = &survey.Input{
	Message: "Enter the URL of the safe",
	Help:    "The URL must be in the format <protocol>://<path>",
}

var urlParam = assist.Param{
	Use:   "url",
	Short: "The URL of the safe",
	Complete: func(c *assist.Command, arg string, params map[string]string) {
		for _, s := range []string{"sftp://", "s3://", "local:///"} {
			if strings.HasPrefix(s, arg) {
				println(s)
			}
		}
	},
	Match: func(c *assist.Command, arg string, args map[string]string) (string, error) {
		if arg == "" {
			err := survey.AskOne(askUrl, &arg)
			if err != nil {
				return "", err
			}
		}
		arg = strings.TrimSpace(arg)
		data, err := core.DecodeBinary(arg)
		if err == nil {
			data, err = security.EcDecrypt(Identity, data)
			if err == nil {
				arg = string(data)
			}
		}

		u, err := url.Parse(arg)
		if err != nil {
			return "", err
		}
		switch u.Scheme {
		case "sftp", "s3", "file":
			break
		default:
			return "", core.Errorf("Invalid URL scheme: %s", u.Scheme)
		}
		return arg, nil
	},
}

var addCmd = &assist.Command{
	Use:    "add",
	Short:  "Add a safe",
	Params: []assist.Param{urlParam},
	Run: func(args map[string]string) error {
		url := args["url"]

		s, err := stash.Open(DB, Identity, url)
		if err != nil {
			return err
		}
		s.Close()

		err = config.SetConfigValue(DB, SafesDomain, s.ID, s.URL, 0, nil)
		if err != nil {
			return err
		}

		println(styles.UseStyle.Render("Added "+path.Base(s.ID)), " ["+s.ID+"]")
		return nil
	},
}

func init() {
	safeCmd.AddCommand(addCmd)
}
