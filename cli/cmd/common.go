package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/cli/styles"
	"github.com/stregato/mio/lib/config"
	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
)

func listSafes() ([]safeDesc, error) {
	urls, err := config.ListConfigKeys(DB, SafesDomain)
	if err != nil {
		return nil, err
	}

	var safes []safeDesc
	for _, u := range urls {
		s, _, _, ok := config.GetConfigValue(DB, SafesDomain, u)
		if !ok {
			continue
		}

		u2, err := url.Parse(u)
		if err != nil {
			continue
		}
		parts := strings.Split(strings.TrimLeft(u2.Path, "/"), "/")
		if len(parts) < 2 {
			continue
		}
		name := parts[len(parts)-1]
		creator := parts[len(parts)-2]
		creatorId, err := security.NewUserId(creator)
		if err != nil {
			continue
		}
		idx := strings.LastIndex(creator, ".")
		if idx > 0 {
			creator = creator[:idx]
		}

		safes = append(safes, safeDesc{
			name:      name,
			creator:   creator,
			creatorId: creatorId,
			url:       s,
		})
	}
	return safes, nil
}

func safeMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	descs, err := listSafes()
	if err != nil {
		return "", err
	}

	names := core.Apply(descs, func(s safeDesc) (string, bool) {
		return s.name, strings.Contains(s.name, arg)
	})

	if len(names) == 0 {
		return "", nil
	}
	if len(names) == 1 {
		return descs[0].url, nil
	}

	description := func(value string, index int) string {
		return descs[index].creator
	}

	var idx int
	err = survey.AskOne(&survey.Select{
		Message:     "Select a safe",
		Options:     names,
		Description: description}, &idx)
	if err != nil {
		return "", err
	}

	return descs[idx].url, nil

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

func matchUser(c *assist.Command, arg string, params map[string]string) (string, error) {
	if arg == "" {
		err := survey.AskOne(&survey.Input{
			Message: "Enter the id of the user to grant access to",
			Help:    "The id must be a valid public key with optional nickname",
		}, &arg)
		if err != nil {
			return "", err
		}
	}
	_, err := security.NewUserId(arg)
	if err != nil {
		return "", err
	}

	return arg, nil
}

var userParam = assist.Param{
	Use:   "user",
	Short: "The id of the user to grant access to",
	Match: matchUser,
}

func matchExisting(c *assist.Command, arg string, params map[string]string) (string, error) {
	s, err := safe.Open(DB, Identity, params["safe"])
	if err != nil {
		return "", err
	}
	defer s.Close()

	groups, err := s.GetGroups()
	if err != nil {
		return "", nil
	}
	users := groups[safe.UserGroup]
	err = survey.AskOne(&survey.Select{
		Message: "Select an existing user",
		Options: core.Apply(users.Slice(), func(u security.ID) (string, bool) {
			v := string(u)
			return v, strings.HasPrefix(v, arg)
		}),
	}, &arg)
	if err != nil {
		return "", err
	}

	return arg, nil
}

func completeExisting(_ *assist.Command, arg string, params map[string]string) {
	s, err := safe.Open(DB, Identity, params["safe"])
	if err != nil {
		return
	}
	defer s.Close()

	groups, err := s.GetGroups()
	if err != nil {
		return
	}
	users := groups[safe.UserGroup]
	for _, u := range users.Slice() {
		if strings.HasPrefix(string(u), arg) {
			println(string(u))
		}
	}
}

var existingParam = assist.Param{
	Use:      "existing",
	Short:    "The id of an existing user",
	Match:    matchExisting,
	Complete: completeExisting,
}

func printGroups(groups safe.Groups) {
	for n, g := range groups {
		fmt.Print(styles.UseStyle.Render(string(n) + ": "))
		for _, u := range g.Slice() {
			fmt.Print(styles.ShortStyle.Render(u.Nick() + " "))
		}
		fmt.Println()
	}
}
