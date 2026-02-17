package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/utils"
	"github.com/serious-snow/govm/pkg/utils/httpc"
	"github.com/serious-snow/govm/pkg/utils/path"
	"github.com/serious-snow/govm/pkg/version"
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
		suggest := suggestVersion(version, ActionInstall)
		if len(suggest) == 0 {
			printError("暂未找到该版本资源下载，请执行：")
			printCmdLine("update")
			return
		}
		installVersion(suggest, force, ignore)
		return
	}

	if err := silentInstall(version, true); err != nil {
		printError(err.Error())
		return
	}

	printInfo("安装成功，如需激活，执行：")
	printCmdLine("use", version)
}

func silentInstall(ver string, checkSha256 bool) error {
	var versionInfo *GoVersionInfo
	version := version.New(ver)
	for _, goVersion := range remoteVersion.Go {
		if version.Equal(goVersion.Version) {
			versionInfo = goVersion
			break
		}
	}
	if versionInfo == nil {
		return errors.New("暂未找到该版本资源下载")
	}

	filename := versionInfo.Filename

	oldSha := versionInfo.Sha256

	if !checkSha256 {
		oldSha = ""
	}

	newFileName := filepath.Join(conf.CachePath, filename)
	download := true
	if path.FileIsExisted(newFileName) {
		if oldSha == "" || utils.CheckSha256(newFileName, oldSha) {
			download = false
		} else {
			if err := os.Remove(newFileName); err != nil {
				return fmt.Errorf("删除损坏的缓存文件失败: %w", err)
			}
		}
	}
	if download {
		Printf("开始下载：%s\n", version.String())
		Printf("下载：%s\n", downloadLink+filename)
		err := httpc.Download(downloadLink+filename, conf.CachePath, filename, oldSha)
		if err != nil {
			return err
		}
	}
	// 然后解压到install文件夹
	toPath := filepath.Join(conf.InstallPath, version.String())
	err := path.Decompress(newFileName, toPath)
	if err != nil {
		_ = os.RemoveAll(toPath)
		return fmt.Errorf("解压失败:%w", err)
	}

	return nil
}
