package commands

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/master/woland/safe"
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/safe"
)

var askUrl = &survey.Input{
	Message: "Enter the URL of the safe",
	Help:    "The URL must be in the format <protocol>://<path>",
}
var askCreator = &survey.Input{
	Message: "Enter the creator id of the safe",
	Help:    "The creator id is the public key of the creator",
}

var add = assist.Command{
	Use:   "add",
	Short: "Add a new safe",
	Params: []assist.Param{
		assist.Param{
			Use: "url",
			Complete: func(c *assist.Command, args []string) (bool, []string) {
				return []string{"sftp://", "s3://", "local:///"}
			},
			Survey: func(c *assist.Command, args []string) string {
				var url string
				err := survey.AskOne(askUrl, &url)
				if err != nil {
					return ""
				}
				return url
			},
		},
		assist.Param{
			Use: "creator",
			Survey: func(c *assist.Command, args []string) string {
				var creatorId string
				err := survey.AskOne(askCreator, &creatorId)
				if err != nil {
					return ""
				}
				return creatorId
			},
		},
	},
	Run: func(c *assist.Command, args []string) {
		url, creator := c.Args["url"], c.Args["creator"]

		s, err := safe.Open(url, creator)
		if err != nil {
			assist.Errorf("Error opening safe: %v", err)
		}
	}

}
