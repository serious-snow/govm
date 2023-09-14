package cmd

import (
	"github.com/urfave/cli/v3"
)

func unholdCommand() *cli.Command {
	return &cli.Command{
		Name:      "unhold",
		Usage:     "Cancel a hold command for a version",
		UsageText: getCmdLine("unhold", "<version>"),
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return cli.ShowSubcommandHelp(c)
			}
			unhold(v)
			return nil
		},
	}
}

func unhold(v string) {
	v = trimVersion(v)
	if !isHold(v) {
		return
	}

	tempHoldVersions := make([]string, 0, len(holdVersions))

	for _, i2 := range holdVersions {
		if i2 == v {
			continue
		}

		tempHoldVersions = append(tempHoldVersions, i2)
	}

	holdVersions = tempHoldVersions

	saveLocalHoldVersion()
}
