package cmd

import (
	"github.com/urfave/cli/v3"
)

func holdCommand() *cli.Command {
	return &cli.Command{
		Name:      "hold",
		Usage:     "Place a version on hold",
		UsageText: getCmdLine("hold", "<version>"),
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return cli.ShowSubcommandHelp(c)
			}
			hold(v)
			return nil
		},
	}
}

func hold(v string) {
	v = trimVersion(v)
	if isHold(v) {
		return
	}
	if !isInInstall(v) {
		printError("该版本未安装")
		return
	}

	holdVersions = append(holdVersions, v)

	saveLocalHoldVersion()
}

func isHold(v string) bool {
	for _, version := range holdVersions {
		if v == version {
			return true
		}
	}
	return false
}
