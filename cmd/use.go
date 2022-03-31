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
		Aliases:   []string{"u"},
		Usage:     "active a <version>",
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

	goRoot := filepath.Join(conf.InstallPath, version, "go")
	if !path.PathIsExisted(goRoot) {
		printError("go 文件夹不存在，请重新安装")
		return
	}

	os.Remove(linkPath)
	err := os.Symlink(goRoot, linkPath)
	if err != nil {
		if isWin && strings.Contains(err.Error(), "A required privilege is not held by the client.") {
			printError("创建软连接失败：没有足够的权限，请使用管理员重试")
			return
		}
		printError("创建软连接失败：" + err.Error())
	}
}
