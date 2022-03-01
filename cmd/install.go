package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/utils/httpc"
	"govm/utils/path"
	"os"
	"path/filepath"
)

func installCommand() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "download and install a <version>.",
		UsageText: getCmdLine("install", "[--force]", "[--ignore-sha256]", "<version>"),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "force install a version",
			},
			&cli.BoolFlag{
				Name:    "ignore-sha256",
				Aliases: []string{"i"},
				Usage:   "ignore check sha256",
			},
		},
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return cli.ShowSubcommandHelp(c)
			}

			installVersion(v, c.Bool("force"), c.Bool("ignore-sha256"))
			return nil
		},
	}
}

func installVersion(version string, force bool, ignore bool) {
	version = trimVersion(version)

	if !force && isInInstall(version) {
		printError(version + " 已经安装，如需覆盖，请执行：")
		if ignore {
			printCmdLine("install", "--force", "--ignore-sha256", version)
		} else {
			printCmdLine("install", "--force", version)
		}
		return
	}

	if !isInLocalCache(version) {
		printError("暂未找到该版本资源下载，请先执行：")
		printCmdLine("list", "--update")
		return
	}

	oldShav := ""

	fileName := getDownloadFilename(version)

	if !ignore {

		b, err := httpc.Get(downloadLink + fileName + ".sha256")
		if err != nil {
			printError("暂未找到该版本sha256资源，请尝试忽略hash校验，执行：")
			if force {
				printCmdLine("install", "--force", "--ignore-sha256", version)
			} else {
				printCmdLine("install", "--ignore-sha256", version)
			}
			return
		}
		oldShav = string(b)
	}

	err := httpc.Download(downloadLink+fileName, conf.CachePath, fileName, oldShav)
	if err != nil {
		printError("\n" + err.Error())
		return
	}
	newFileName := filepath.Join(conf.CachePath, fileName)
	fr, err := os.Open(newFileName)
	if err != nil {
		printError("\n解压失败")
		return
	}
	defer fr.Close()
	//然后解压到install文件夹
	toPath := filepath.Join(conf.InstallPath, version)
	err = path.Decompress(newFileName, toPath)
	if err != nil {
		printError(fmt.Sprint("\n解压失败", err))
		os.RemoveAll(toPath)
		return
	}
	printInfo("\n安装成功，如需激活，执行：")
	printCmdLine("use", version)
}
