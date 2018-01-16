package cmd

import (
	"fmt"

	bcex "github.com/RichardWeiYang/bcex/lib"
	"github.com/jawher/mow.cli"
)

func (c *CLI) RegisterCommands() {
	// list
	c.Command("list", "List Exchanges", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			lists := bcex.ListEx()
			for _, ex := range lists {
				fmt.Println(ex)
			}
		}
	})

	c.Command("balance", "Get Account Balance", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			exs := bcex.GetExs()
			for n, ex := range exs {
				balances, err := ex.GetBalance()
				if err == nil {
					fmt.Println(n + ":")
					for i, b := range balances {
						fmt.Println(i, b.Currency, b.Balance)
					}
				} else {
					fmt.Println(err)
				}
			}
		}
	})
}
