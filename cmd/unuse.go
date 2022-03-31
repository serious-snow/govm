package cmd

import (
	"github.com/urfave/cli/v2"
	"os"
)

func unuseCommand() *cli.Command {
	return &cli.Command{
		Name:      "unuse",
		Aliases:   []string{"uu"},
		Usage:     "deactivated current use version",
		UsageText: getCmdLine("unuse"),
		Action: func(c *cli.Context) error {
			if !currentUse.Valid() {
				printError("当前没有激活的版本")
				return nil
			}
			os.Remove(linkPath)
			return nil
		},
	}
}
