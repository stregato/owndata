package cmd

import (
	"fmt"
	"slices"
	"time"

	"github.com/stregato/stash/cli/assist"
	"github.com/stregato/stash/cli/styles"
	"github.com/stregato/stash/lib/db"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/stregato/stash/lib/stash"
)

func clearLines(lines int) {
	if lines < 1 {
		return // If the number of lines is less than 1, do nothing.
	}

	// Move the cursor up 'lines' lines
	fmt.Printf("\033[%dA", lines)

	// Clear each line
	for i := 0; i < lines; i++ {
		fmt.Print("\033[K") // Clear the current line
		if i < lines-1 {
			fmt.Print("\033[1B") // Move the cursor down one line, if not the last iteration
		}
	}

	// Optionally move the cursor back to the original position after clearing
	fmt.Printf("\033[%dA", lines) // Move up again to where clearing started
}

func listenRun(params map[string]string) error {

	s, err := getSafeByName(params["safe"])
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := db.Open(s, stash.UserGroup, nil)
	if err != nil {
		return err
	}
	defer d.Close()

	var lines []string
	for {
		_, err = d.Sync()
		if err != nil {
			return err
		}
		rows, err := d.Query("GET_MESSAGES", sqlx.Args{"limit": 10})
		if err != nil {
			return err
		}

		for rows.Next() {
			var message string
			var createdAt time.Time
			var creatorId security.ID
			var contentType string

			err = rows.Scan(&message, &createdAt, &creatorId, &contentType)
			if err != nil {
				continue
			}

			if contentType != "text/plain" {
				lines = append(lines, styles.ErrorStyle.Render("Unsupported content type: "+contentType))
			} else {
				pre := fmt.Sprintf("%s %s:", createdAt.Format("15:04"), creatorId.Nick())
				lines = append(lines, styles.UseStyle.Render(pre)+styles.ShortStyle.Render(message))
			}
		}

		slices.Reverse(lines)
		for _, line := range lines {
			fmt.Println(line)
		}

		time.Sleep(5 * time.Second)
		clearLines(len(lines))
		lines = nil
	}
}

var listenCmd = &assist.Command{
	Use:   "listen",
	Short: "Listen for incoming messages",
	Params: []assist.Param{
		safeParam,
	},
	Run: listenRun,
}

func init() {
	chatCmd.AddCommand(listenCmd)
}
