package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/utils/filepath"
	"os"
	"path"
	"runtime"
)

func uninstallCommand() *cli.Command {
	return &cli.Command{
		Name:      "uninstall",
		Aliases:   nil,
		Usage:     "uninstall a <version>.",
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
		tarFile := path.Join(conf.CachePath, fmt.Sprintf("go%s.%s-%s.tar.gz", version, runtime.GOOS, runtime.GOARCH))
		if filepath.FileIsExisted(tarFile) {
			os.Remove(tarFile)
		}
	}()
	err := os.RemoveAll(path.Join(conf.InstallPath, version))
	if err != nil {
		printError(version + " 卸载失败：" + err.Error())
		return
	}

	fmt.Println(version, "卸载成功")
}
