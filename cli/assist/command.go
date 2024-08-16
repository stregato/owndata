package assist

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stregato/stash/cli/styles"
	"github.com/stregato/stash/lib/core"
)

var (
	Completion *bool
	Echo       *bool
)

type Param struct {
	Use string // Use is  one-line usage message.

	Short string // Short is the short description shown in the 'help' output.

	Complete func(c *Command, arg string, params map[string]string) // Complete is the function to call to complete the argument in bash.

	Match func(c *Command, arg string, params map[string]string) (string, error) // Survey is the function to call to get the argument from the user.

	Multiple bool // Multiple is true if the parameter can be repeated. Only last parameter can be repeated.
}

type Command struct {
	Use string // Use is the one-line usage message.

	Short string // Short is the short description shown in the 'help' output.

	Params []Param // List of parameters

	Subcommands []*Command // List of subcommands

	Help func(matched int) // Help is the function to call when the command is executed with no arguments.

	Match func(arg string) bool // Match is the function to call to check if the arguments match the command.

	Run func(args map[string]string) error // Run is the function to call when the command is executed.

	Parent *Command // Parent is the parent command
}

func BashQuote(input string) string {
	// Define the special characters that require the string to be quoted
	specialChars := ` |&;()<>{}$*?!"'~#^\=` // Include backslash and both quotes for completeness

	// Check if any special character is in the input
	for _, char := range specialChars {
		if strings.ContainsRune(input, char) {
			// If a special character is found, return the string with double quotes
			return `"` + strings.ReplaceAll(input, `"`, `\"`) + `"`
		}
	}

	// If no special character is found, return the input unchanged
	return input
}

func (c *Command) Execute() {

	var oks = []string{"✅", "😊", "👍", "🉑"}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

repeat:
	n := c
	i := 0

	interactive := false
	var args = append([]string{}, flag.Args()...)
	var echo = filepath.Base(os.Args[0])

	// println()
	// for _, a := range args {
	// 	println("[", a, "]")
	// }

	for {
		for ; i < len(args); i++ {
			// Find the subcommand that matches the argument
			idx := findMatch(args[i], n.Subcommands)
			if idx >= 0 {
				n = n.Subcommands[idx]
				echo += " " + args[i]
			} else {
				break
			}
		}
		// If we have a subcommand, continue to the next argument
		if len(n.Subcommands) == 0 {
			break
		}

		if *Completion {
			var filter string
			if i < len(args) {
				filter = strings.ToLower(args[i])
			}
			for _, s := range n.Subcommands {
				if strings.HasPrefix(s.Use, filter) {
					fmt.Println(s.Use, " ")
				}
			}

			return
		}

		var cmd string
		options := core.Apply(n.Subcommands, func(c *Command) (string, bool) {
			return c.Use, true
		})

		if n.Parent != nil {
			options = append(options, "back")
		} else {
			options = append(options, "exit")
		}
		description := func(value string, index int) string {
			if index >= len(n.Subcommands) {
				switch value {
				case "exit":
					return "Exit the program"
				case "back":
					return "Go back to the previous command"
				}
			}
			c := n.Subcommands[index]
			return styles.ShortStyle.Render(c.Short)
		}
		interactive = true

		// Ask the user to select a subcommand
		err := survey.AskOne(&survey.Select{
			Message:     "Select a command",
			Options:     options,
			Description: description,
		}, &cmd)
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(err.Error()))
			return
		}

		switch cmd {
		case "exit":
			os.Exit(0)
		case "back":
			n = n.Parent
			i--
		default:
			// Add the selected subcommand to the arguments
			args = append(args, strings.TrimRight(cmd, " "))
			echo += " " + cmd
		}
	}

	j := 0
	params := map[string]string{}
	l := len(args)
	// Parse the parameters
	for j < len(n.Params) {
		var v string
		if i < l {
			v = args[i]
			i++
		}

		p := n.Params[j]

		if *Completion {
			if p.Complete != nil {
				p.Complete(n, v, params)
			}
			return
		}

		// try to match the parameter
		v, err := p.Match(n, v, params)
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(err.Error()))
			n.help(args, i)
			return
		}
		core.Info("matched arg %s with param %s, final value %s", v, p.Use, v)
		params[p.Use] = v
		echo += " " + BashQuote(v)
		// If the parameter is multiple and there are more arguments, try to match the next argument
		if p.Multiple && i < len(args) {
			continue
		}
		j++
	}

	// Check if there are arguments that are not matched
	if i < len(args) {
		fmt.Println(styles.ErrorStyle.Render("Too many arguments"))
		n.help(args, i)
	}

	// If the command has no Run function, show the help
	if n.Run == nil {
		n.help(args, i)
		return
	}
	core.Info("running command %s with params %v", n.Use, params)
	// Run the command
	err := n.Run(params)
	if err != nil {
		fmt.Println(styles.ErrorStyle.Render(err.Error()))
		os.Exit(1)
	} else {
		println(oks[r.Intn(len(oks))])
	}

	if *Echo {
		println("cmd: ", echo)
	}

	if interactive {
		goto repeat
	}
}

func findMatch(token string, commands []*Command) int {
	for i, cmd := range commands {
		if cmd.Match != nil && cmd.Match(token) {
			return i
		}
		if cmd.Use == token {
			return i
		}
	}
	return -1
}

func (c *Command) help(args []string, matched int) {
	// If the command has a Help function, call it
	if c.Help != nil {
		c.Help(matched)
		return
	}

	var parts []string
	parts = append(parts, filepath.Base(os.Args[0]))
	parts = append(parts, args...)
	m := strings.Join(parts, " ")

	// if the command has no subcommands or parameters, print the usage
	if len(c.Subcommands) > 0 {
		fmt.Println("Usage: " + styles.UsageStyle.Render(m+" [command]"))
	}

	// if the command has parameters, print the usage
	if len(c.Params) > 0 {
		ps := core.Apply(c.Params, func(p Param) (string, bool) {
			return fmt.Sprintf("[%s]", p.Use), true
		})
		fmt.Println("Usage: " + styles.UsageStyle.Render(m+" "+strings.Join(ps, " ")))
	}

	c.printSubcommands()
	c.printParams()
}

func (c *Command) printSubcommands() {
	if len(c.Subcommands) > 0 {
		fmt.Println("")
		fmt.Println("with command")

		var shorts = map[string]string{}
		var uses []string
		for _, cmd := range c.Subcommands {
			uses = append(uses, cmd.Use)
			shorts[cmd.Use] = cmd.Short
		}

		for _, use := range uses {
			u := styles.UseStyle.Render(use)
			short := styles.UsageStyle.Render(shorts[use])
			fmt.Println(u + " " + short)
		}
	}
}

func (c *Command) printParams() {
	if len(c.Params) > 0 {
		fmt.Println("")
		fmt.Println("with parameters")

		for _, param := range c.Params {
			use := styles.UseStyle.Render(param.Use)
			short := styles.UsageStyle.Render(param.Short)
			fmt.Println(use + " " + short)
		}
	}
}

func (c *Command) AddCommand(cmd *Command) {
	// Add the subcommand to the list of subcommands
	c.Subcommands = append(c.Subcommands, cmd)
	cmd.Parent = c
}
