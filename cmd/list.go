package cmd

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/utils/httpc"
	"github.com/serious-snow/govm/pkg/version"
	"github.com/serious-snow/govm/types"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"l"},
		Usage:     "Show version list",
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
		printError("列表更新失败：" + err.Error())
		return
	}

	Println("列表更新完成,本次更新 新增数量为:", len(res)-len(localCacheVersions))

	localCacheVersions = res
	saveLocalCacheVersion()
}

func getAvailable() ([]*version.Version, error) {
	//https://storage.googleapis.com/golang/?prefix=go&marker=
	link := downloadLink + "?prefix=go&marker="
	suffix := "tar\\.gz"
	if isWin {
		suffix = "zip"
	}
	var (
		buf     []byte
		err     error
		res     []*version.Version
		result  types.ListBucketResult
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
				res = append(res, version.New(reg.FindStringSubmatch(content.Key)[1]))
			}
		}
		if result.NextMarker == "" {
			break
		}

		newLink = link + result.NextMarker

		result.Reset()
	}
	version.SortV(res).Reverse()
	return res, nil
}

func printAvailable() {
	sb := strings.Builder{}
	using, other := "->     ", "       "
	for _, v := range localCacheVersions {
		if isInstall(*v) {
			if version.Equal(*v, currentUse) {
				_, _ = color.New(color.FgGreen).Fprint(&sb, using, v.String())
			} else {
				_, _ = color.New(color.FgBlue).Fprint(&sb, other, v.String())
			}

		} else {
			sb.WriteString(other)
			sb.WriteString(v.String())
		}
		sb.WriteString("\n")
	}
	Print(sb.String())
	sb.Reset()
}

func printInstalled() {
	sb := strings.Builder{}
	using, other := "->     ", "       "
	for _, v := range localInstallVersion {
		if version.Equal(*v, currentUse) {
			_, _ = color.New(color.FgGreen).Fprint(&sb, using, v.String())
		} else {
			_, _ = color.New(color.FgBlue).Fprint(&sb, other, v.String())
		}
		sb.WriteString("\n")
	}
	Print(sb.String())
	sb.Reset()
}
