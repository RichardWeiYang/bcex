package cmd

import (
	"fmt"

	//bcex "github.com/RichardWeiYang/bcex/lib"
	"github.com/jawher/mow.cli"
)

func (c *CLI) RegisterCommands() {
	// list
	c.Command("list", "List Exchanges", func(cmd *cli.Cmd) {
		fmt.Println("None")
	})
}
