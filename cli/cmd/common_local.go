package cmd

import (
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/mio/cli/assist"
)

func mountMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	u, _ := url.Parse(params["safe"])

	if arg == "" {
		desktopFolder, err := getDesktopFolder()
		if err != nil {
			return "", err
		}
		arg = filepath.Join(desktopFolder, path.Base(u.Path))
	}
	return arg, nil
}

var mountParam = assist.Param{
	Use:   "mount",
	Short: "The path to mount the safe on",
	Match: mountMatch,
}

func dirMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		ls, err := os.ReadDir(dir)
		if err != nil {
			return "", err
		}
		dirs := []string{"✔"}
		if dir != "/" {
			dirs = append(dirs, "◄")
		}
		for _, f := range ls {
			if f.IsDir() {
				dirs = append(dirs, f.Name())
			}
		}

		println("Current directory:", dir)
		var d string
		err = survey.AskOne(&survey.Select{Message: "Select a sub directory or confirm current",
			Options: dirs}, &d)
		if err != nil {
			return "", err
		}
		switch d {
		case "✔":
			return dir, nil
		case "◄":
			dir, _ = path.Split(dir)
		default:
			dir = path.Join(dir, d)
		}
	}
}

var dirParam = assist.Param{
	Use:   "dir",
	Short: "a local directory",
	Match: dirMatch,
}
