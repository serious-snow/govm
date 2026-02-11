package cmd

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/utils/path"
)

func useCommand() *cli.Command {
	return &cli.Command{
		Name:      "use",
		Aliases:   []string{"u"},
		Usage:     "Active a <version>",
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
		suggest := suggestVersion(version, ActionUse)
		if suggest == "" {
			printError("该版本未安装，请先安装，执行：")
			printCmdLine("install", version)
			return
		}
		version = suggest
	}
	goRoot := filepath.Join(conf.InstallPath, version, "go")
	if !path.PathIsExisted(goRoot) {
		printError("go 文件夹不存在，请重新安装")
		return
	}

	_ = os.Remove(linkPath)
	err := Symlink(goRoot, linkPath)
	if err != nil {
		printError("创建软连接失败：" + err.Error())
	}
}
