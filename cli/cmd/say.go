package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/db"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/stregato/stash/lib/stash"
)

var messageParam = assist.Param{
	Use:   "message",
	Short: "The message to send",
	Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
		if arg == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Message:",
				Help:    "The message to send",
			}, &arg)
			if err != nil {
				return "", err
			}
		}
		return arg, nil
	},
}

func writeMessage(p db.Database, message string) error {
	_, err := p.Exec("INSERT_MESSAGE", sqlx.Args{"message": message,
		"createdAt":   core.Now(),
		"creatorId":   Identity.Id,
		"contentType": "text/plain",
	})
	if err != nil {
		return err
	}
	_, err = p.Sync()
	return err
}

func sayRun(args map[string]string) error {
	message := args["message"]

	s, err := getSafeByName(args["safe"])
	if err != nil {
		return err
	}

	p, err := db.Open(s, stash.UserGroup, nil)
	if err != nil {
		return err
	}

	err = writeMessage(p, message)
	if err != nil {
		return err
	}

	println("Ok. Manuscripts don't burn!")

	return nil
}

var sayCmd = &assist.Command{
	Use:   "say",
	Short: "Say something in a chat",

	Params: []assist.Param{safeParam, messageParam},
	Run:    sayRun,
}

func init() {
	chatCmd.AddCommand(sayCmd)
}
