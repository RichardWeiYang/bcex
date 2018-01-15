package cmd

import "github.com/jawher/mow.cli"

// CLI struct for main
type CLI struct {
	*cli.Cli
}

// NewCLI initializes new command line interface
func NewCLI() *CLI {
	c := &CLI{cli.App("bcex", "A BlockChain Exchange CLI")}

	return c
}
