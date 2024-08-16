package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"path"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/fs"
	"github.com/stregato/stash/lib/stash"
)

func getSafeByName(name string) (*stash.Stash, error) {
	safes, err := listSafes()
	if err != nil {
		return nil, err
	}

	for _, s := range safes {
		if s.name == name {
			return stash.Open(DB, Identity, s.url)
		}
	}
	return nil, core.Errorf("Safe %s not found", name)
}

func getSafeAndPath(name string) (*stash.Stash, string, error) {
	dir := ""
	if firstSlash := strings.Index(name, "/"); firstSlash > 0 {
		dir = name[firstSlash+1:]
		name = name[:firstSlash]
	}

	s, err := getSafeByName(name)
	if err != nil {
		return nil, "", err
	}

	return s, dir, nil
}

func safeMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	descs, err := listSafes()
	if err != nil {
		return "", err
	}

	descs = core.Apply(descs, func(s safeDesc) (safeDesc, bool) {
		return s, strings.Contains(s.name, arg)
	})

	if len(descs) == 0 {
		return "", nil
	}
	if len(descs) == 1 {
		if getSafeName(descs[0].url) == arg {
			return arg, nil
		}
	}

	var options []string
	count := map[string]int{}

	for _, s := range descs {
		count[s.name]++
		options = append(options, fmt.Sprintf("%s by %s [%d]", s.name, s.creator, count[s.name]))
	}

	var idx int
	err = survey.AskOne(&survey.Select{
		Message: "Select a safe",
		Options: options}, &idx)
	if err != nil {
		return "", err
	}

	if count[descs[idx].name] > 1 {
		return descs[idx].url, nil
	} else {
		return descs[idx].name, nil
	}
}

func safeComplete(_ *assist.Command, arg string, _ map[string]string) {
	safes, err := listSafes()
	if err == nil {
		for _, s := range safes {
			if strings.Contains(s.url, arg) {
				println(s.url)
			}
		}
	}
}

var safeParam = assist.Param{
	Use:      "safe",
	Short:    "The safe to use",
	Match:    safeMatch,
	Complete: safeComplete,
}

type pathMatchOptions struct {
	msg      string
	safePath bool
	onlyDir  bool
	onlyFile bool
}

func selectFile(file string, f *fs.FileSystem, options pathMatchOptions) (string, error) {
	if f == nil && file == "" {
		file, _ = os.Getwd()
	}

	prompt := &survey.Input{Message: options.msg,
		Suggest: func(toComplete string) []string {
			if f == nil {
				files, _ := filepath.Glob(toComplete + "*")
				return files
			}

			ls, _ := f.List(toComplete, fs.ListOptions{})
			var files []string
			for _, l := range ls {
				if l.IsDir || !options.onlyDir {
					files = append(files, l.Name)
				}
			}
			return files
		},
		Default: file,
	}
	err := survey.AskOne(prompt, &file)
	if err != nil {
		return "", err
	}
	return file, nil
}

func pathMatch(options pathMatchOptions) func(_ *assist.Command, arg string, _ map[string]string) (string, error) {
	return func(_ *assist.Command, arg string, _ map[string]string) (string, error) {
		var f *fs.FileSystem
		var safeName, file string
		var err error

		if options.safePath {
			safeName, file = arg, ""
			if firstSlash := strings.Index(arg, "/"); firstSlash > 0 {
				safeName = arg[:firstSlash]
				file = arg[firstSlash+1:]
			}

			safeName, err = safeMatch(nil, safeName, nil)
			if err != nil {
				return "", err
			}

			s, err := getSafeByName(safeName)
			if err != nil {
				return "", err
			}
			defer s.Close()

			f, err = fs.Open(s)
			if err != nil {
				return "", err
			}
			defer f.Close()
		} else {
			file = arg
		}

		file, err = selectFile(file, f, options)
		if err != nil {
			return "", err
		}

		return path.Join(safeName, file), nil
	}
}

func pathComplete(options pathMatchOptions) func(_ *assist.Command, arg string, _ map[string]string) {
	return func(_ *assist.Command, arg string, _ map[string]string) {
		var f *fs.FileSystem
		var safeName string

		if options.safePath {
			safeName, _ = arg, ""
			if firstSlash := strings.Index(arg, "/"); firstSlash > 0 {
				safeName = arg[:firstSlash]
				arg = arg[firstSlash+1:]
			}
			sds, err := listSafes()
			if err != nil {
				return
			}

			candidates := []string{}
			for _, s := range sds {
				if s.name == safeName {
					candidates = append(candidates, s.name)
					break
				}
				if strings.Contains(s.name, arg) {
					candidates = append(candidates, s.name)
				}
			}

			switch len(candidates) {
			case 0:
				return
			case 1:
				safeName = candidates[0]
			default:
				for _, c := range candidates {
					fmt.Println(c + "/")
				}
				return
			}
			s, err := getSafeByName(safeName)
			if err != nil {
				return
			}
			defer s.Close()

			f, err = fs.Open(s)
			if err != nil {
				return
			}
			defer f.Close()

			var dir string
			idx := strings.LastIndex(arg, "/")
			if idx >= 0 {
				dir = arg[:idx+1]
				arg = arg[idx+1:]
			}

			ls, _ := f.List(dir, fs.ListOptions{})
			for _, l := range ls {
				if options.onlyDir && !l.IsDir {
					continue
				}
				if options.onlyFile && l.IsDir {
					continue
				}
				if strings.HasPrefix(l.Name, arg) {
					if l.IsDir {
						fmt.Println(assist.BashQuote(path.Join(safeName, l.Name) + "/"))
					} else {
						fmt.Println(assist.BashQuote(path.Join(safeName, l.Name)))
					}
				}
			}
			return
		}

		ls, _ := os.ReadDir(arg)
		for _, l := range ls {
			if options.onlyDir && !l.IsDir() {
				continue
			}

			if l.IsDir() {
				fmt.Println(assist.BashQuote(l.Name() + "/"))
			} else {
				fmt.Println(assist.BashQuote(l.Name()))
			}
		}
	}
}
