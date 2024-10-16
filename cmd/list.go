package cmd

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
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
		UsageText: getCmdLine("list", "[--installed]"),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "installed",
				Aliases: []string{"i"},
				Usage:   "show installed version list",
			},
			&cli.BoolFlag{
				Name:    "upgradeable",
				Aliases: []string{"u"},
				Usage:   "show upgradeable version list",
			},
		},

		Action: func(c *cli.Context) error {

			if len(remoteVersion.GoVersions) == 0 {
				reloadAvailable()
			}

			switch {
			case c.Bool("installed"):
				printInstalled()
			case c.Bool("upgradeable"):
				printUpgradeable()
			default:
				printAvailable()
			}
			return nil

		},
	}
}

func reloadAvailable() {
	Println("正在拉取 go 最新版本列表...")
	spin := spinner.New(spinner.CharSets[14], time.Millisecond*100)
	spin.Start()
	res, err := getAvailable()
	if err != nil {
		spin.Stop()
		printError("列表更新失败：" + err.Error())
		return
	}
	spin.Stop()

	if len(localInstallVersions) != 0 {
		Println("列表更新完成, 本次更新 新增数量为:", len(res)-len(remoteVersion.GoVersions))
	}

	remoteVersion.GoVersions = res
	saveLocalRemoteVersion()
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
	printVersions(remoteVersion.GoVersions)
}

func printInstalled() {
	printVersions(localInstallVersions)
}

func printVersions(vs []*version.Version) {
	sb := strings.Builder{}
	using, other, holding := "->     ", "       ", " (hold)"
	for _, v := range vs {
		if isInstall(*v) {
			holdStr := ""
			if isHold(v.String()) {
				holdStr = holding
			}
			if version.Equal(*v, currentUse) {
				_, _ = color.New(color.FgGreen).Fprint(&sb, using, v.String(), holdStr)
			} else {
				_, _ = color.New(color.FgBlue).Fprint(&sb, other, v.String(), holdStr)
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

func printUpgradeable() {
	m := getUpgradeableList()
	if len(m) == 0 {
		println("所有版本均是最新")
		return
	}

	sb := strings.Builder{}
	count := 0
	sbHold := strings.Builder{}
	holdCount := 0
	for s, versions := range m {
		for _, v := range versions {
			if isHold(v.String()) {
				holdCount++
				sbHold.WriteString(v.String())
				sbHold.WriteString(" -> ")
				sbHold.WriteString(s)
				sbHold.WriteString("\n")
				continue
			}
			count++
			sb.WriteString(v.String())
			sb.WriteString(" -> ")
			sb.WriteString(s)
			sb.WriteString("\n")
		}
	}

	Printf("%d 个有最新版本，其中：\n可升级：%d 个\n被保留：%d 个\n", count+holdCount, count, holdCount)

	if count != 0 {
		Println("可升级的版本：")
		Print(sb.String())
	}
	if holdCount != 0 {
		Println("被保留的版本：")
		Print(sbHold.String())
	}
}
