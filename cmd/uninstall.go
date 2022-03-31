package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/utils/path"
	"os"
	"path/filepath"
)

func uninstallCommand() *cli.Command {
	return &cli.Command{
		Name:      "uninstall",
		Aliases:   []string{"ui"},
		Usage:     "uninstall a <version>",
		UsageText: getCmdLine("uninstall", "<version>"),
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return cli.ShowSubcommandHelp(c)
			}
			uninstallVersion(v)
			return nil
		},
	}
}

func uninstallVersion(version string) {

	version = trimVersion(version)

	if !isInInstall(version) {
		printError(version + " 未安装")
		return
	}
	defer func() {
		fileName := getDownloadFilename(version)

		if path.FileIsExisted(fileName) {
			os.Remove(fileName)
		}
	}()
	err := os.RemoveAll(filepath.Join(conf.InstallPath, version))
	if err != nil {
		printError(version + " 卸载失败：" + err.Error())
		return
	}

	fmt.Println(version, "卸载成功")
}
