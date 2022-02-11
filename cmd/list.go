package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/models"
	"govm/utils/httpc"
	"regexp"
	"runtime"
	"strings"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   nil,
		Usage:     "show list.",
		UsageText: getCmdLine("list", "[--available]", "[--update]"),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "available",
				Aliases: []string{"a"},
				Usage:   "show available version list",
			},
			&cli.BoolFlag{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "update available version list",
			},
		},

		Action: func(c *cli.Context) error {
			if c.Bool("update") {
				reloadAvailable()
				if c.Bool("available") {
					printAvailable()
				}
				return nil
			}
			if c.Bool("available") {
				if len(localCacheVersions) == 0 {
					reloadAvailable()
				}
				//打印本地列表
				printAvailable()
				return nil
			}
			//打印已安装的 根据目录检测
			printInstalled()
			return nil

		},
	}
}

func reloadAvailable() {
	res, err := getAvailable()
	if err != nil {
		printError("更新列表失败：" + err.Error())
		return
	}

	fmt.Println("更新列表完成,本次更新 新增数量为:", len(res)-len(localCacheVersions))

	localCacheVersions = res
	saveLocalCacheVersion()
}

func getAvailable() ([]*models.Version, error) {
	//https://storage.googleapis.com/golang/?prefix=go&marker=
	link := downloadLink + "?prefix=go&marker="
	suffix := "tar\\.gz"
	if isWin {
		suffix = "zip"
	}
	var (
		buf     []byte
		err     error
		res     []*models.Version
		result  models.ListBucketResult
		goOS    = runtime.GOOS
		goArch  = runtime.GOARCH
		reg     = regexp.MustCompile(fmt.Sprintf("^go(.*)\\.%s-%s\\.%s$", goOS, goArch, suffix))
		newLink = link
	)
	for {
		buf, err = httpc.Get(newLink)
		if err != nil {
			return nil, err
		}
		err = xml.Unmarshal(buf, &result)
		if err != nil {
			return nil, err
		}

		for _, content := range result.Contents {
			if reg.MatchString(content.Key) {
				res = append(res, models.NewVInfo(reg.FindStringSubmatch(content.Key)[1]))
			}
		}
		if result.NextMarker == "" {
			break
		}

		newLink = link + result.NextMarker

		result.Reset()
	}
	models.SortV(res).Reverse()
	return res, nil
}

func printAvailable() {
	sb := strings.Builder{}
	sb.WriteString("available list:\n\n")
	for _, i2 := range localCacheVersions {

		if isInstall(*i2) {
			sb.WriteString("\033[1;32m")
			sb.WriteString(i2.String())
			sb.WriteString(" (installed)")
			sb.WriteString("\033[0m")
		} else {
			sb.WriteString(i2.String())
		}
		sb.WriteString("\n")
	}
	fmt.Println(sb.String())
	sb.Reset()
}

func printInstalled() {
	sb := strings.Builder{}
	sb.WriteString("installed list:\n\n")
	sb.WriteString("\033[1;32m")
	for _, i2 := range localInstallVersion {
		if i2.Compare(currentUse) == 0 {
			sb.WriteString("\033[1;31m")
			sb.WriteString(i2.String())
			sb.WriteString(" (current use)")
			sb.WriteString("\033[1;32m")
		} else {
			sb.WriteString(i2.String())
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\033[0m")
	fmt.Println(sb.String())
	sb.Reset()
}
