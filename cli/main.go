/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"
	"path/filepath"

	"github.com/stregato/mio/cli/cmd"
	"github.com/stregato/mio/cli/commands"
	"github.com/stregato/mio/cli/styles"
	"github.com/stregato/mio/lib/sqlx"
)

// getDBPath returns the path to the database file built from the application folder, the folder 'mio' and the file 'mio.db'.
func getDBPath() (string, error) {
	// Get the user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Construct the path to the database file
	dbPath := filepath.Join(configDir, "mio", "mio.db")

	return dbPath, nil
}

func openDB() error {
	// Get the path to the database file
	dbPath, err := getDBPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create the directory if it does not exist
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Open the database file
	db, err := sqlx.Open(dbPath)
	if err != nil {
		return err
	}
	commands.DB = db

	return nil
}

func main() {

	if err := openDB(); err != nil {
		styles.ErrorStyle.Render("Error opening database: ", err.Error())
	}

	cmd.Execute()
}
