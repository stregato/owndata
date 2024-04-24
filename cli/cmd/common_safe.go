package cmd


func getSafeByNameOrUrl(nameOrUrl string) (*safe.Safe, error) {
	safes, err := listSafes()
	if err != nil {
		return nil, err
	}

	for _, s := range safes {
		if s.name == nameOrUrl || s.url == nameOrUrl {
			return safe.Open(DB, Identity, s.url)
		}
	}
	return nil, core.Errorf("Safe %s not found", nameOrUrl)
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

func completeSafe(_ *assist.Command, arg string, _ map[string]string) {
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
	Complete: completeSafe,
}



func safePathMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	safeName, dir := arg, ""
	if firstSlash := strings.Index(arg, "/"); firstSlash > 0 {
		safeName = arg[:firstSlash]
		dir = arg[firstSlash+1:]
	}

	safeName, err := safeMatch(c, safeName, params)
	if err != nil {
		return "", err
	}

	s, err := getSafeByNameOrUrl(safeName)
	if err != nil {
		return "", err
	}

	for {
		ls, err := s.ListDir(dir, safe.ListOptions{})
		if err != nil {
			return "", err
		}
		dirs := core.Apply(ls, func(f safe.File) (string, bool) {
			if f.IsDir() {
				return f.Name(), true
			}
			return "", false
		})
		dirs = append([]string{"✔", "◄"}, dirs...)

		var d string
		err = survey.AskOne(&survey.Select{Message: "Select a directory", Options: dirs}, &d)
		if err != nil {
			return "", err
		}
		switch d {
		case "✔":
			break
		case "◄":
			dir, _ = path.Split(dir)
		default:
			dir = path.Join(dir, d)
	}

	return path.Join(safeName, dir), nil
}

var safeDirParam = assist.Param{
	Use:   "safeDir",
	Short: "A dir in the safe",
	Match: safePathMatch,
}
