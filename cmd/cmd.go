package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/config"
	"github.com/serious-snow/govm/pkg/utils/path"
	"github.com/serious-snow/govm/pkg/version"
)

const (
	downloadLink = "https://go.dev/dl/"
)

var (
	app *cli.Command

	flagNoSuggest bool

	Version = "dev"
)

var (
	conf                 config.Config
	homeDir              string
	processDir           string
	remoteVersion        RemoteVersion
	localInstallVersions []*version.Version
	holdVersions         []string

	currentUse version.Version

	pName    string
	isWin    bool
	linkPath string // 软连接路径
	envPath  string // 环境变量路径
)

func init() {
	pName = filepath.Base(os.Args[0])
	isWin = runtime.GOOS == "windows"
}

func Run() error {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		return err
	}
	processDir = filepath.Join(homeDir, ".govm")

	err = path.MakeDir(processDir)
	if err != nil {
		return err
	}

	{
		configPath := filepath.Join(processDir, "conf.yaml")
		conf, err = config.InitConfig(processDir, configPath)
		if err != nil {
			return err
		}

		err = path.MakeDir(conf.InstallPath)
		if err != nil {
			return err
		}

		err = path.MakeDir(conf.CachePath)
		if err != nil {
			return err
		}
	}

	linkPath = filepath.Join(processDir, "go")
	envPath = filepath.Join(linkPath, "bin")

	{
		app = &cli.Command{
			Name:        pName,
			Usage:       "Manage go version",
			UsageText:   "",
			ArgsUsage:   "",
			Version:     Version,
			Description: "a go version manager.\n" + printEnv(),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:       "no-colors",
					Usage:      "disable colors",
					Persistent: true,
				},
				&cli.BoolFlag{
					Name:        "no-suggest",
					Usage:       "disable suggest",
					Persistent:  true,
					Destination: &flagNoSuggest,
				},
			},
			Before: func(c *cli.Context) error {
				color.NoColor = color.NoColor || c.Bool("no-colors")
				return nil
			},

			Commands: []*cli.Command{
				listCommand(),
				installCommand(),
				useCommand(),
				cacheCommand(),
				uninstallCommand(),
				unuseCommand(),
				execCommand(),
				updateCommand(),
				upgradeCommand(),
				holdCommand(),
				unholdCommand(),
			},
			UseShortOptionHandling: true,
			Suggest:                true,
			Reader:                 os.Stdin,
			Writer:                 os.Stdout,
			ErrWriter:              os.Stderr,
		}

		sort.Sort(cli.FlagsByName(app.Flags))
		sort.Slice(app.Commands, func(i, j int) bool {
			return app.Commands[i].Name < app.Commands[j].Name
		})
	}

	{
		// 检查环境变量
		initEnvPath()
		// 读取本地安装版本
		readLocalInstallVersion()
		// 读取本地缓存列表
		readLocalRemoteVersion()
		// 读取当前使用的版本
		readCurrentUseVersion()
		// 读取本地hold的版本列表
		readLocalHoldVersion()
	}

	return app.Run(context.Background(), os.Args)
}

func printEnv() string {
	sb := strings.Builder{}

	if !strings.Contains(os.Getenv("PATH"), envPath) {
		sb.WriteString(fmt.Sprintf("\nplease set environment：%s", color.RedString(envPath)))
	} else {
		sb.WriteString("\nenvironment set success.")
	}

	if len(remoteVersion.Govm.Version) == 0 || Version == "dev" {
		return sb.String()
	}

	remote := version.New(remoteVersion.Govm.Version)
	local := version.New(Version)

	if !version.Equal(*remote, *local) {
		sb.WriteString(fmt.Sprintf("\nplease update govm to %s", color.GreenString(remoteVersion.Govm.Version)))
	}

	return sb.String()
}

func isInstall(info version.Version) bool {
	for _, vInfo := range localInstallVersions {
		if vInfo.Compare(info) == 0 {
			return true
		}
	}
	return false
}

func readLocalRemoteVersion() {
	cacheJsonPath := filepath.Join(conf.CachePath, "version.json")
	buf, err := os.ReadFile(cacheJsonPath)
	if err != nil {
		return
	}
	// Try to unmarshal into full struct first
	if err := json.Unmarshal(buf, &remoteVersion); err != nil {
		// If that fails, try unmarshaling into just the Go field for backward compatibility
		_ = json.Unmarshal(buf, &remoteVersion.Go)
	}
}

func saveLocalRemoteVersion() error {
	cacheJsonPath := filepath.Join(conf.CachePath, "version.json")
	file, err := os.OpenFile(cacheJsonPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(remoteVersion); err != nil {
		return err
	}
	return nil
}

func readLocalHoldVersion() {
	filename := filepath.Join(conf.CachePath, "hold.json")
	buf, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	holdVersions = make([]string, 0)
	_ = json.Unmarshal(buf, &holdVersions)
}

func saveLocalHoldVersion() error {
	filename := filepath.Join(conf.CachePath, "hold.json")
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(holdVersions); err != nil {
		return err
	}
	return nil
}

func readLocalInstallVersion() {
	fileInfoList, err := os.ReadDir(conf.InstallPath)
	if err != nil {
		return
	}
	localInstallVersions = make([]*version.Version, 0)
	for _, info := range fileInfoList {
		if !info.IsDir() {
			continue
		}

		vInfo := version.New(info.Name())
		if vInfo.Valid() {
			localInstallVersions = append(localInstallVersions, vInfo)
		}
	}

	version.SortV(localInstallVersions).Reverse()
}

func printError(msg string) {
	ErrorLn(color.RedString(msg))
}

func printInfo(msg string) {
	Println(color.GreenString(msg))
}

func isInLocalCache(ver string) bool {
	v := version.New(ver)
	for _, cacheVersion := range remoteVersion.Go {
		if version.Equal(*v, cacheVersion.Version) {
			return true
		}
	}
	return false
}

func isInInstall(ver string) bool {
	v := version.New(ver)
	for _, cacheVersion := range localInstallVersions {
		if version.Equal(*v, *cacheVersion) {
			return true
		}
	}

	return false
}

func readCurrentUseVersion() {
	to, err := os.Readlink(linkPath)
	if err != nil {
		return
	}

	// Use filepath to properly handle path separators
	relPath, err := filepath.Rel(conf.InstallPath, to)
	if err != nil {
		return
	}

	// Clean the path and extract version
	versionStr := filepath.Base(relPath)
	versionStr = strings.TrimPrefix(versionStr, "go")
	v := version.New(versionStr)
	if v != nil {
		currentUse = *v
	}
}

func trimVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "go")
	version = strings.TrimPrefix(version, "v")
	return version
}

func getCmdLine(cmd ...string) string {
	newCmd := append([]string{pName}, cmd...)
	return strings.Join(newCmd, " ")
}

func printCmdLine(cmd ...string) {
	Println(getCmdLine(cmd...))
}

func formatSize(size int64) string {
	fSize := float64(size)
	units := []string{"B", "KB", "MB", "GB"}
	idx := 0
	for fSize >= 1024 && idx < len(units)-1 {
		fSize /= 1024
		idx++
	}
	return fmt.Sprintf("%.2f%s", fSize, units[idx])
}

func initEnvPath() {
	if strings.Contains(os.Getenv("PATH"), envPath) {
		return
	}

	if conf.AutoSetEnv != nil {
		return
	}

	prompt := promptui.Prompt{
		Label:     "是否设置环境变量",
		IsConfirm: true,
		Default:   "y",
	}

	v, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) {
			os.Exit(130)
		}
	}

	ok := strings.ToLower(v) == "y"

	conf.AutoSetEnv = &ok
	conf.Sync()

	if ok {
		SetEnv()
	}
}

func getDownloadFilename(version string) string {
	suffix := "tar.gz"
	if isWin {
		suffix = "zip"
	}

	return fmt.Sprintf("go%s.%s-%s.%s", version, runtime.GOOS, runtime.GOARCH, suffix)
}

type Action uint16

const (
	ActionInstall Action = iota + 1
	ActionUse
	ActionExec
)

func suggestVersion(ver string, action Action) string {
	if flagNoSuggest {
		return ""
	}
	v := version.New(ver)

	minorVersion := v.MinorVersion()
	if v.String() != minorVersion {
		return ""
	}
	ver = ""
	switch action {
	case ActionInstall:
		versions := make([]*version.Version, 0, len(remoteVersion.Go))
		for _, info := range remoteVersion.Go {
			versions = append(versions, &info.Version)
		}
		vls := GetMinorGroup(versions)[minorVersion]
		for _, ve := range vls {
			if isInInstall(ve.String()) {
				ver = ve.String()
				break
			}
		}
		if ver == "" {
			if len(vls) != 0 {
				ver = vls[0].String()
			}
		}

	case ActionUse, ActionExec:
		vls := GetMinorGroup(localInstallVersions)[minorVersion]
		for _, ve := range vls {
			if isInInstall(ve.String()) {
				ver = ve.String()
				break
			}
		}
	}

	return ver
}

func GetMinorGroup(list []*version.Version) map[string][]*version.Version {
	m := make(map[string][]*version.Version)
	for _, v := range list {
		m[v.MinorVersion()] = append(m[v.MinorVersion()], v)
	}
	return m
}
