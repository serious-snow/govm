package cmd

import (
	"github.com/urfave/cli/v2"
	"govm/utils/path"
	"os"
	"path/filepath"
	"strings"
)

func useCommand() *cli.Command {
	return &cli.Command{
		Name:      "use",
		Aliases:   nil,
		Usage:     "active a <version>.",
		UsageText: getCmdLine("use", "<version>"),
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return cli.ShowSubcommandHelp(c)
			}
			useVersion(v)
			return nil
		},
	}
}

func useVersion(version string) {

	version = trimVersion(version)

	if !isInInstall(version) {
		printError("该版本未安装，请先安装，执行：")
		printCmdLine("install", version)
		return
	}

	goBinPath := filepath.Join(conf.InstallPath, version, "go", "bin")
	if !path.PathIsExisted(goBinPath) {
		printError("golang bin 文件夹不存在，请重新安装")
		return
	}

	os.Remove(linkPath)
	err := os.Symlink(goBinPath, linkPath)
	if err != nil {
		if isWin && strings.Contains(err.Error(), "A required privilege is not held by the client.") {
			printError("创建软连接失败：没有足够的权限，请使用管理员重试")
			return
		}
		printError("创建软连接失败：" + err.Error())
	}
}
