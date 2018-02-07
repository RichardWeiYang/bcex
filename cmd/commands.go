package cmd

import (
	"fmt"
	"strconv"

	. "github.com/RichardWeiYang/bcex/lib"
	"github.com/jawher/mow.cli"
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

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
			exchanges := ListEx()
			for _, n := range exchanges {
				if *exname != "all" && n != *exname {
					continue
				}
				ex := GetEx(n)
				ek := keys[n]
				ex.SetKey(ek.AccessKeyId, ek.SecretKeyId)
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

			ek := keys[*exname]
			ex.SetKey(ek.AccessKeyId, ek.SecretKeyId)
			cp := NewCurrencyPair2(*currencypair)
			price, err := ex.GetPrice(&cp)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				fmt.Printf("%0.8f\n", price.Price)
			}
		}
	})

	c.Command("symbols", "Get supported symbols", func(cmd *cli.Cmd) {
		var (
			exname = cmd.StringArg("EX", "bigone", "The Exchange to query")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			ex := GetEx(*exname)
			if ex == nil {
				fmt.Println(*exname, ": not supported")
				return
			}

			symbols, err := ex.GetSymbols()
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				for _, s := range symbols {
					fmt.Printf("%s ", s)
				}
				fmt.Println()
			}
		}
	})

	c.Command("depth", "Get depth for currency pair", func(cmd *cli.Cmd) {
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
			depth, err := ex.GetDepth(&cp)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				fmt.Println("Depth of ", *currencypair, "on ", *exname)
				fmt.Println("\tPrice      \tAmount")
				fmt.Println("Asks:")
				for i := min(5, len(depth.Asks)); i >= 1; i-- {
					fmt.Printf("\t%0.8f\t%0.8f\n",
						depth.Asks[len(depth.Asks)-i].Price,
						depth.Asks[len(depth.Asks)-i].Amount)
				}
				fmt.Println("Bids:")
				for i := 0; i < min(5, len(depth.Bids)); i++ {
					fmt.Printf("\t%0.8f\t%0.8f\n",
						depth.Bids[i].Price,
						depth.Bids[i].Amount)
				}
			}
		}
	})

	c.Command("neworder", "place an order", func(cmd *cli.Cmd) {
		var (
			exname       = cmd.StringArg("EX", "bigone", "The Exchange to query")
			side         = cmd.StringArg("SD", "sell/buy", "buy or sell")
			currencypair = cmd.StringArg("CP", "btc_usd", "CurrencyPair to query(lower case)")
			price        = cmd.StringArg("PI", "0.01", "The price you want to buy or sell")
			amount       = cmd.StringArg("AM", "0.2", "The amount you want to buy or sel")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			ex := GetEx(*exname)
			if ex == nil {
				fmt.Println(*exname, ": not supported")
				return
			}

			ek := keys[*exname]
			ex.SetKey(ek.AccessKeyId, ek.SecretKeyId)

			cp := NewCurrencyPair2(*currencypair)
			price_f, _ := strconv.ParseFloat(*price, 64)
			amount_f, _ := strconv.ParseFloat(*amount, 64)
			order := Order{
				CP:     cp,
				Side:   *side,
				Price:  price_f,
				Amount: amount_f,
			}

			id, err := ex.NewOrder(&order)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				fmt.Println("ID:      ", id)
			}
		}
	})

	c.Command("cancelorder", "cancel an order", func(cmd *cli.Cmd) {
		var (
			symbol = cmd.StringOpt("s symbol", "", "order symbol if necessary")
			exname = cmd.StringArg("EX", "bigone", "The Exchange to query")
			id     = cmd.StringArg("ID", "id", "order id")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			ex := GetEx(*exname)
			if ex == nil {
				fmt.Println(*exname, ": not supported")
				return
			}

			ek := keys[*exname]
			ex.SetKey(ek.AccessKeyId, ek.SecretKeyId)

			o := Order{Id: *id, CP: NewCurrencyPair2(*symbol)}
			err := ex.CancelOrder(&o)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Done")
			}
		}
	})

	c.Command("queryorder", "query an order", func(cmd *cli.Cmd) {
		var (
			symbol = cmd.StringOpt("s symbol", "", "order symbol if necessary")
			exname = cmd.StringArg("EX", "bigone", "The Exchange to query")
			id     = cmd.StringArg("ID", "id", "order id")
		)

		cmd.Action = func() {
			Init(*bcexKey)
			ex := GetEx(*exname)
			if ex == nil {
				fmt.Println(*exname, ": not supported")
				return
			}

			ek := keys[*exname]
			ex.SetKey(ek.AccessKeyId, ek.SecretKeyId)

			order := Order{Id: *id, CP: NewCurrencyPair2(*symbol)}
			o, err := ex.QueryOrder(&order)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("ID:      ", o.Id)
				fmt.Println("Symbol:  ", o.CP.String())
				fmt.Println("Side:    ", o.Side)
				fmt.Printf("Price:    %0.8f\n", o.Price)
				fmt.Printf("Amount:   %0.8f\n", o.Amount)
				fmt.Printf("Executed: %0.8f\n", o.Executed)
				fmt.Printf("Remain:   %0.8f\n", o.Remain)
				fmt.Println("State:   ", o.State)
			}
		}
	})
}
