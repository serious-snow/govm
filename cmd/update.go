package cmd

import (
	"github.com/urfave/cli/v3"
)

func updateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update available version list",
		UsageText: getCmdLine("update"),
		Action: func(c *cli.Context) error {
			checkGovmUpdate(c.Context)
			reloadAvailable()
			printCanUpgradeCount()
			return nil
		},
	}
}

func printCanUpgradeCount() {
	m := getUpgradeableList()
	canUpgradeCount := 0

	for _, versions := range m {
		canUpgradeCount += len(versions)
	}

	if canUpgradeCount == 0 {
		return
	}

	Printf("%d 个版本有最新版本, 执行 %s 查看更多信息\n", canUpgradeCount, getCmdLine("list --upgradeable"))
}
