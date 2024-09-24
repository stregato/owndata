package cmd

import (
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
)

var Root = &assist.Command{
	Use:   "stash",
	Short: "stash is a CLI tool to manage encrypted safes on remote servers",
}

var (
	DB       *sqlx.DB
	Identity *security.Identity
	DBPath   *string
	Loglevel *string
)

//go:embed cli1_0.sql
var ddl string

var askNick = &survey.Input{
	Message: "Enter your nickname to create a new identity",
	Help:    "This nickname will be bound to your public key and used to identify you with other users",
}

func setupDB() {

	var err error
	if *DBPath == "" {
		//retrieve user's application directory using os.UserConfigDir
		dir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}
		*DBPath = filepath.Join(dir, "safe.db")
	} else {
		println("Using database at ", *DBPath)
	}

	//create a new sqlx.DB instance
	DB, err = sqlx.Open(*DBPath)
	if err != nil {
		panic(err)
	}

	// try to get the current identity from the database with config.GetConfigStruct
	err = config.GetConfigStruct(DB, config.SettingsDomain, "identity", &Identity)
	if err == sqlx.ErrNoRows {
		var nick string

		err = survey.AskOne(askNick, &nick)
		if err != nil {
			panic(err)
		}

		nick = strings.Trim(nick, " ")
		if nick == "" {
			panic("nickname cannot be empty")
		}

		// if no identity is found, create a new one
		Identity = security.NewIdentityMust(nick)
		err = config.SetConfigStruct(DB, config.SettingsDomain, "identity", Identity)
		if err != nil {
			panic(err)
		}
		println(Identity.Id)

	}

	err = DB.Define(1.0, ddl)
	if err != nil {
		panic(err)
	}

}

func setFlags() {
	assist.Completion = flag.Bool("completion", false, "generate bash completion script")
	//	assist.Completion = true
	DBPath = flag.String("db", "", "path to the database file")
	Loglevel = flag.String("log", "error", "log level")
	assist.Echo = flag.Bool("echo", false, "echo the command")

	flag.Parse()

	switch *Loglevel {
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		panic("invalid log level")
	}
}

func init() {
	setFlags()
	setupDB()
}
