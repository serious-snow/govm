package cmd

import (
	"os"

	"github.com/urfave/cli/v3"
)

func unuseCommand() *cli.Command {
	return &cli.Command{
		Name:      "unuse",
		Aliases:   []string{"uu"},
		Usage:     "Deactivated current use version",
		UsageText: getCmdLine("unuse"),
		Action: func(c *cli.Context) error {
			if !currentUse.Valid() {
				printError("当前没有激活的版本")
				return nil
			}
			if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
				printError("删除软连接失败：" + err.Error())
				return nil
			}
			return nil
		},
	}
}
