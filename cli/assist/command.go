package assist

import (
	"os"
	"slices"
	"strings"

	"github.com/stregato/mio/cli/styles"
)

type Param struct {
	Use string // Use is  one-line usage message.

	Complete func(c *Command, args []string) (bool, []string) // Complete is the function to call to complete the argument in bash.

	Survey func(c *Command, args []string) string // Survey is the function to call to get the argument from the user.
}

type Command struct {
	Use string // Use is the one-line usage message.

	Short string // Short is the short description shown in the 'help' output.

	Params []Param // List of parameters

	Subcommands []Command // List of subcommands

	Help func(matched int) // Help is the function to call when the command is executed with no arguments.

	Match func(arg string) bool // Match is the function to call to check if the arguments match the command.

	Run func(c *Command, args []string) // Run is the function to call when the command is executed.

	args map[string]string // Args is the list of arguments to the command.

	parent *Command // Parent is the parent command.
}

func (c *Command) Execute() {
	n := c
	for i := 1; i < len(os.Args); i++ {
		idx := findMatch(os.Args[i], n.Subcommands)
		if idx == -1 {
			n.Help(i - 1)
			return
		}
		line = line + " " + os.Args[i]
		n = &n.Subcommands[idx]
	}
}

func findMatch(token string, commands []Command) int {
	for i, cmd := range commands {
		if cmd.Use == token {
			return i
		}
	}
	return -1
}

func (c *Command) getUsageLine() string {
	var parts []string

	for n := c; n != nil; n = n.parent {
		parts = append(parts, n.Use)
	}
	parts = slices.Reverse(parts)
	if c.Subcommands != nil {
		parts = append(parts, "[command]")
		return strings.Join(parts, " ")
	}

}

func (c *Command) Help(matched int) {
	if c.Help != nil {
		c.Help(matched)
		return
	}

	m := strings.Join(os.Args[0:matched], " ")
	if len(c.Subcommands) > 0 {
		styles.UsageStyle.Render(m + " [command]")
	}
	if len(c.Params) > 0 {
		core.
			styles.UsageStyle.Render(m + " [command]")
	}

}

func (c *Command) AddCommand(cmd Command) {
	c.Subcommands = append(c.Subcommands, cmd)
}
