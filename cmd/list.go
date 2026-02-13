package cmd

import (
	"encoding/json"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/utils/httpc"
	"github.com/serious-snow/govm/pkg/version"
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
			if len(remoteVersion.Go) == 0 {
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
		Println("列表更新完成, 本次更新 新增数量为:", len(res)-len(remoteVersion.Go))
	}

	remoteVersion.Go = res
	if err := saveLocalRemoteVersion(); err != nil {
		printError("保存版本列表失败：" + err.Error())
	}
}

func getAvailable() ([]*GoVersionInfo, error) {
	// https://go.dev/dl/?mode=json&include=all
	link := downloadLink + "?mode=json&include=all"

	var result []*ListGoVersionResponse
	buf, err := httpc.Get(link)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(buf, &result); err != nil {
		return nil, err
	}
	list := make([]*GoVersionInfo, 0, len(result))

	seen := map[string]struct{}{}
	for _, response := range result {
		for _, file := range response.Files {

			vv := trimVersion(file.Version)
			if file.Kind != "archive" {
				continue
			}
			if file.Os != runtime.GOOS || file.Arch != runtime.GOARCH {
				continue
			}
			if _, ok := seen[vv]; ok {
				continue
			}
			seen[vv] = struct{}{}

			list = append(list, &GoVersionInfo{
				Filename: file.Filename,
				Sha256:   file.Sha256,
				Size:     file.Size,
				Version:  *version.New(vv),
			})
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Version.Greater(list[j].Version)
	})

	return list, nil
}

func printAvailable() {
	versions := make([]*version.Version, 0, len(remoteVersion.Go))
	for _, v := range remoteVersion.Go {
		versions = append(versions, &v.Version)
	}
	printVersions(versions)
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
