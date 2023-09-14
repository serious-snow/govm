package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/utils"
	"github.com/serious-snow/govm/pkg/utils/httpc"
	"github.com/serious-snow/govm/pkg/utils/path"
)

func installCommand() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Aliases:   []string{"i"},
		Usage:     "Download and install a <version>",
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

	oldSha := ""

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
		oldSha = string(b)
	}

	if err := silentInstall(version, oldSha); err != nil {
		printError(err.Error())
		return
	}

	printInfo("安装成功，如需激活，执行：")
	printCmdLine("use", version)
}

func silentInstall(version string, oldSha string) error {

	fileName := getDownloadFilename(version)

	newFileName := filepath.Join(conf.CachePath, fileName)
	download := true
	if path.FileIsExisted(newFileName) {
		if utils.CheckSha256(newFileName, oldSha) {
			download = false
		} else {
			_ = os.Remove(newFileName)
		}
	}
	if download {
		Printf("开始下载：%s\n", version)
		err := httpc.Download(downloadLink+fileName, conf.CachePath, fileName, oldSha)
		if err != nil {
			return err
		}
	}
	// 然后解压到install文件夹
	toPath := filepath.Join(conf.InstallPath, version)
	err := path.Decompress(newFileName, toPath)
	if err != nil {
		_ = os.RemoveAll(toPath)
		return fmt.Errorf("解压失败:%w", err)
	}

	return nil
}
