package cmd

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
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

func pathMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
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

var pathParam = assist.Param{
	Use:   "path",
	Short: "The path to mount the safe on",
	Match: pathMatch,
}

func mountRun(params map[string]string) error {
	url := params["safe"]
	path := params["path"]

	s, err := safe.Open(DB, Identity, url)
	if err != nil {
		return err
	}
	defer s.Close()

	err = mountFS(s, path)
	if err != nil {
		return err
	}

	return nil
}

var mountCmd = &assist.Command{
	Use:    "mount",
	Short:  "Mount a safe on the local filesystem",
	Params: []assist.Param{safeParam, pathParam},
	Run:    mountRun,
}

func init() {
	if runtime.GOOS == "linux" {
		Root.AddCommand(mountCmd)
	}
}
