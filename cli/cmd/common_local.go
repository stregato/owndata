//go:build linux
// +build linux

package cmd

import (
	"net/url"
	"path"
	"path/filepath"

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
