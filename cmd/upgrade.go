package cmd

import (
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/version"
)

func upgradeCommand() *cli.Command {
	return &cli.Command{
		Name:      "upgrade",
		Usage:     "Upgrade outdated version list",
		UsageText: getCmdLine("upgrade"),
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			switch v {
			case "govm":
				upgradeGOVM(c.Context)
				return nil
			default:
				if v != "" {
					upgrade(v)
					return nil
				}
				upgradeAll()
				return nil
			}
		},
	}
}

func upgrade(v string) {
	v = trimVersion(v)
	if !isInInstall(v) {
		printError(v + " 未安装")
		return
	}
	if isHold(v) {
		printError(v + " 被标记为保留")
		return
	}

	current := *version.New(v)

	newest := getPatchNewestVersion(current)

	if newest == nil {
		printError(v + " 找不到最新版本")
		return
	}

	if version.Equal(current, *newest) {
		Println(current, "已经最新版本")
		return
	}

	upgradeVersions(map[string][]*version.Version{
		newest.String(): {
			&current,
		},
	})

}

func upgradeAll() {
	m := getUpgradeableList()
	if len(m) == 0 {
		Println("没有可升级的版本")
		return
	}
	sb := strings.Builder{}
	for s, versions := range m {
		for _, v := range versions {
			if isHold(v.String()) {
				continue
			}
			sb.WriteString(v.String())
			sb.WriteString(" -> ")
			sb.WriteString(s)
			sb.WriteString("\n")
		}
	}

	if len(sb.String()) != 0 {
		Println("升级版本：")
		Print(sb.String())
	}
	upgradeVersions(m)
}

func upgradeVersions(m map[string][]*version.Version) {

	var (
		installCount   int
		uninstallCount int
		ignoreCount    int
	)

	for s, versions := range m {

		canInstall := false
		for _, v := range versions {
			if !isHold(v.String()) {
				canInstall = true
				break
			}
		}

		if !canInstall {
			ignoreCount += len(versions)
			continue
		}

		if !isInInstall(s) {
			if err := silentInstall(s, ""); err != nil {
				printError(err.Error())
				continue
			}
			installCount++
		}

		for _, v := range versions {
			if isHold(v.String()) {
				ignoreCount++
				continue
			}

			uninstallVersion(v.String())
			uninstallCount++

			// 如果卸载的是当前正在使用的 就设置为刚刚的最新版本
			if version.Equal(currentUse, *v) {
				readLocalInstallVersion()
				useVersion(s)
				readCurrentUseVersion()
			}
		}
	}

	if installCount+uninstallCount != 0 {
		readLocalInstallVersion()
	}

	Printf("共升级 %d 个版本, 安装了 %d 个版本, 卸载了 %d 个版本，忽略了 %d 个版本\n", len(m), installCount, uninstallCount, ignoreCount)

}

func getPatchNewestVersion(v version.Version) *version.Version {
	for _, v2 := range remoteVersion.GoVersions {
		if v2.Major == v.Major && v2.Minor == v.Minor {
			return v2
		}
	}
	return nil
}

//获取可以更新的列表

func getUpgradeableList() map[string][]*version.Version {
	result := make(map[string][]*version.Version)

	m := GetPatchNewestVersionMap()

	for _, v := range localInstallVersions {
		mVer := v.MinorVersion()

		latest := m[mVer]
		if latest == nil {
			continue
		}

		if version.Equal(*latest, *v) {
			continue
		}

		result[latest.String()] = append(result[latest.String()], v)
	}
	return result
}

// 1.18->1.18.10
func GetPatchNewestVersionMap() map[string]*version.Version {
	result := make(map[string]*version.Version)
	for _, cacheVersion := range remoteVersion.GoVersions {
		old := cacheVersion.MinorVersion()
		if _, ok := result[old]; !ok {
			result[old] = cacheVersion
		}
	}
	return result
}
