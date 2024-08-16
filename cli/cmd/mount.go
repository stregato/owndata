//go:build linux
// +build linux

package cmd

import (
	"os"
	"path"
	"runtime"

	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/fs"
)

func getDesktopFolder() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Typically, the Desktop folder on Windows can be accessed via the USERPROFILE environment variable
		return os.Getenv("USERPROFILE") + "\\Desktop", nil
	case "darwin":
		// On macOS, the Desktop directory is usually under the HOME directory
		return os.Getenv("HOME") + "/Desktop", nil
	case "linux":
		// On Linux, the Desktop directory is also typically in the HOME directory
		return os.Getenv("HOME") + "/Desktop", nil
	default:
		// Fallback for any other operating system not explicitly mentioned
		return "", core.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func mountRun(params map[string]string) error {
	safe := params["safe"]

	s, err := getSafeByName(safe)
	if err != nil {
		return err
	}
	defer s.Close()

	ph, err := getDesktopFolder()
	if err != nil {
		return err
	}
	ph = path.Join(ph, safe)

	f, err := fs.Open(s)
	if err != nil {
		return err
	}

	err = f.Mount(ph)
	if err != nil {
		return err
	}

	return nil
}

var mountCmd = &assist.Command{
	Use:    "mount",
	Short:  "Mount a safe on the local filesystem",
	Params: []assist.Param{safeParam, mountParam},
	Run:    mountRun,
}

func init() {
	if runtime.GOOS == "linux" {
		Root.AddCommand(mountCmd)
	}
}
