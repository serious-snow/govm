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
		for _, version := range versions {
			if !isHold(version.String()) {
				canUpgradeCount++
			}
		}
	}

	if canUpgradeCount == 0 {
		return
	}

	Printf("有 %d 个版本可以升级到最新, 执行 %s 查看更多信息\n", canUpgradeCount, getCmdLine("list --upgradeable"))
}
