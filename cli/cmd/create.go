package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/lib/config"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
)

var tyInput = &survey.Select{
	Message: "Enter the type of storage",
	Help:    "The type of storage can be s3, local, or sftp",
	Options: []string{"s3", "local", "sftp", "generic"},
}

var s3qs = []*survey.Question{
	{
		Name:     "host",
		Prompt:   &survey.Input{Message: "Enter the S3 host?"},
		Validate: survey.Required,
	},
	{
		Name:     "bucket",
		Prompt:   &survey.Input{Message: "Enter the S3 bucket:"},
		Validate: survey.Required,
	},
	{
		Name:     "accessKey",
		Prompt:   &survey.Input{Message: "Enter the access key?"},
		Validate: survey.Required,
	},
	{
		Name:     "secret",
		Prompt:   &survey.Password{Message: "Enter the secret?"},
		Validate: survey.Required,
	},
}

var genericqs = []*survey.Question{
	{
		Name:     "url",
		Prompt:   &survey.Input{Message: "Enter the url of the store?"},
		Validate: survey.Required,
	},
}

var sftpqs = []*survey.Question{
	{
		Name:     "host",
		Prompt:   &survey.Input{Message: "Enter the SFTP host?"},
		Validate: survey.Required,
	},
	{
		Name:     "path",
		Prompt:   &survey.Input{Message: "Enter the SFTP path:"},
		Validate: survey.Required,
	},
	{
		Name:     "user",
		Prompt:   &survey.Input{Message: "Enter the SFTP user:"},
		Validate: survey.Required,
	},
	{
		Name:     "password",
		Prompt:   &survey.Password{Message: "Enter the SFTP password?"},
		Validate: survey.Required,
	},
}

var localqs = []*survey.Question{
	{
		Name:     "path",
		Prompt:   &survey.Input{Message: "Enter the local path?"},
		Validate: survey.Required,
	},
}

var qs = map[string][]*survey.Question{
	"s3":      s3qs,
	"local":   localqs,
	"sftp":    sftpqs,
	"generic": genericqs,
}

type answers struct {
	Host      string
	Bucket    string
	AccessKey string
	Secret    string
	Url       string
	Path      string
	User      string
	Password  string
}

func storeMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	if arg != "" {
		_, err := url.Parse(arg)
		if err == nil {
			return arg, nil
		}
	}

	var result string
	for {
		var typ string
		err := survey.AskOne(tyInput, &typ)
		if err != nil {
			return "", err
		}

		var as answers
		err = survey.Ask(qs[typ], &as)
		if err != nil {
			return "", err
		}

		as.Host = strings.Trim(as.Host, " ")
		as.Bucket = strings.Trim(as.Bucket, " ")
		as.AccessKey = strings.Trim(as.AccessKey, " ")
		as.Secret = strings.Trim(as.Secret, " ")
		as.Url = strings.Trim(as.Url, " ")
		as.Path = strings.Trim(as.Path, " ")

		switch typ {
		case "s3":
			result = fmt.Sprintf("s3://%s/%s?a=%s&s=%s", as.Host, as.Bucket,
				as.AccessKey, as.Secret)
		case "local":
			result = fmt.Sprintf("local://%s", as.Path)
		case "sftp":
			result = fmt.Sprintf("sftp://%s:%s@%s/%s", as.User, as.Password, as.Host, as.Path)
		case "generic":
			result = as.Url
		}

		var confirm bool
		err = survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("Do you want to use %s as the store?", result)}, &confirm)
		if err != nil {
			return "", err
		}
		if confirm {
			break
		}
	}
	return result, nil
}

var storeParam = assist.Param{
	Use:   "store",
	Short: "The url of the store",
	Complete: func(c *assist.Command, arg string, params map[string]string) {
		for _, s := range []string{"sftp://", "s3://", "local:///"} {
			if strings.HasPrefix(s, arg) {
				println(s)
			}
		}
	},
	Match: storeMatch,
}

func creatorMatch(c *assist.Command, arg string, params map[string]string) (string, error) {
	if arg != "" {
		_, err := security.NewUserId(arg)
		if err != nil {
			return "", err
		}
		return arg, nil
	}

	var creatorId string
	input := &survey.Input{
		Message: "Enter the creator public id:",
		Default: string(Identity.Id),
	}
	err := survey.AskOne(input, &creatorId)
	if err != nil {
		return "", err
	}
	_, err = security.NewUserId(creatorId)
	if err != nil {
		return "", err
	}
	return creatorId, nil
}

var creatorParam = assist.Param{
	Use:   "creator",
	Short: "The creator publid id of the safe",
	Complete: func(c *assist.Command, arg string, params map[string]string) {
		println(Identity.Id.String())
	},
	Match: creatorMatch,
}

var nameParam = assist.Param{
	Use:   "name",
	Short: "The name of the safe",
	Match: func(c *assist.Command, arg string, params map[string]string) (string, error) {
		if arg == "" {
			var name string
			err := survey.AskOne(&survey.Input{Message: "Enter the name of the safe:"}, &name)
			if err != nil {
				return "", err
			}
			return name, nil
		}
		return arg, nil
	},
}

var createCmd = &assist.Command{
	Use:    "create",
	Short:  "create a new safe",
	Params: []assist.Param{storeParam, creatorParam, nameParam},
	Run: func(args map[string]string) error {
		s, err := safe.Create(DB, Identity, args["store"], args["name"])
		if err != nil {
			return err
		}
		_, err = s.UpdateGroup(safe.AdminGroup, safe.Grant, Identity.Id)
		if err != nil {
			return err
		}
		groups, err := s.UpdateGroup(safe.UserGroup, safe.Grant, Identity.Id)
		if err != nil {
			return err
		}

		s.Close()
		err = config.SetConfigValue(DB, SafesDomain, s.ID, s.URL, 0, nil)
		if err != nil {
			return err
		}

		fmt.Println("Safe created successfully. Url: ", s.URL)
		printGroups(groups)

		return nil
	},
}

func init() {
	safeCmd.AddCommand(createCmd)
}
