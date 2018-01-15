package main

import (
	"os"

	"github.com/RichardWeiYang/bcex/cmd"
)

func main() {
	cli := cmd.NewCLI()
	cli.RegisterCommands()
	cli.Run(os.Args)
}
