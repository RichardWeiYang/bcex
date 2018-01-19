package cmd

import (
	"github.com/jawher/mow.cli"
)

var bcexKey *string

// CLI struct for main
type CLI struct {
	*cli.Cli
}

// NewCLI initializes new command line interface
func NewCLI() *CLI {
	c := &CLI{cli.App("bcex", "A BlockChain Exchange CLI")}

	bcexKey = c.String(cli.StringOpt{
		Name:      "k bcex-key",
		Desc:      "BCEX Key",
		EnvVar:    "BCEX_KEY",
		HideValue: true,
	})

	return c
}
