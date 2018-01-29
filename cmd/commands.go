package cmd

import (
	"fmt"

	. "github.com/RichardWeiYang/bcex/lib"
	"github.com/jawher/mow.cli"
)

func (c *CLI) RegisterCommands() {
	// list
	c.Command("list", "List Exchanges", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			lists := ListEx()
			for _, ex := range lists {
				fmt.Println(ex)
			}
		}
	})

	c.Command("setkey", "Set Exchange API-KEY", func(cmd *cli.Cmd) {
		var (
			exname    = cmd.StringArg("EX", "", "The Exchange to set")
			accesskey = cmd.StringArg("AK", "", "The Exchange to set")
			secretkey = cmd.StringArg("SK", "", "The Exchange to set")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			WriteConf(*exname, *accesskey, *secretkey)
		}
	})

	c.Command("balance", "Get Account Balance", func(cmd *cli.Cmd) {
		var (
			exname = cmd.StringArg("EX", "all", "The Exchange Name to display")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			exs := GetExs()
			for n, ex := range exs {
				if *exname != "all" && n != *exname {
					continue
				}
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

	c.Command("price", "Get current price", func(cmd *cli.Cmd) {
		var (
			exname       = cmd.StringArg("EX", "bigone", "The Exchange to query")
			currencypair = cmd.StringArg("CP", "btc_usd", "CurrencyPair to query(lower case)")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			ex := GetEx(*exname)
			if ex == nil {
				fmt.Println(*exname, ": not supported")
				return
			}

			cp := NewCurrencyPair2(*currencypair)
			price, err := ex.GetPrice(&cp)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				fmt.Printf("%0.8f\n", price.Price)
			}
		}
	})
}
