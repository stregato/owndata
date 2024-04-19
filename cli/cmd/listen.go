package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/stregato/mio/cli/assist"
	"github.com/stregato/mio/cli/styles"
	"github.com/stregato/mio/lib/db"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
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

	s, err := safe.Open(DB, Identity, params["safe"])
	if err != nil {
		return err
	}

	d, err := db.Open(s, nil, safe.UserGroup)
	if err != nil {
		return err
	}

	var linesCnt int
	keyPressChan := make(chan rune)
	go func(keyPressChan chan rune) {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Fatalf("failed to read key: %v", err)
				os.Exit(1)
			}
			if key == keyboard.KeyEsc {
				fmt.Println("Exiting...")
				os.Exit(0)
			}
			keyPressChan <- char

		}
	}(keyPressChan)

	for {
		_, err = d.Sync()
		if err != nil {
			return err
		}
		rows, err := d.Query("GET_MESSAGES", sqlx.Args{"limit": 20})
		if err != nil {
			return err
		}

		clearLines(linesCnt)
		linesCnt = 0
		for rows.Next() {
			var message string
			var createdAt time.Time
			var creatorId security.ID

			err = rows.Scan(&message, &createdAt, &creatorId)
			if err != nil {
				continue
			}

			pre := fmt.Sprintf("%s %s:", createdAt.Format("15:04"), creatorId.Nick())
			println(styles.UseStyle.Render(pre), styles.ShortStyle.Render(message))
			linesCnt++
		}
		println()

		var msg string
		select {
		case key := <-keyPressChan:
			if key == '\n' {
				err = writeMessage(d, msg)
				if err != nil {
					return err
				}
				msg = ""
			} else {
				msg += string(key)
			}
		case <-time.After(5 * time.Second):
		}
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
	Root.AddCommand(listenCmd)
}
