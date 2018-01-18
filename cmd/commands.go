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
				fmt.Println(n + ":")
				if err == nil {
					if len(balances) == 0 {
						fmt.Println("\tNone")
						continue
					}
					for _, b := range balances {
						fmt.Println("\t", b.Currency, b.Balance)
					}
				} else {
					fmt.Println("\tError:" + err.Error())
				}
			}
		}
	})
}
