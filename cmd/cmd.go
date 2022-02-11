package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/config"
	"govm/models"
	"govm/utils/filepath"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
)

const (
	downloadLink = "https://storage.googleapis.com/golang/"
)

var (
	conf                config.Config
	homeDir             string
	processDir          string
	localCacheVersions  []*models.Version
	localInstallVersion []*models.Version
	currentUse          models.Version

	pName    string
	isWin    bool
	linkPath string
)

func Run() error {
	pName = path.Base(os.Args[0])
	isWin = runtime.GOOS == "windows"

	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		return err
	}
	processDir = path.Join(homeDir, ".govm")

	err = filepath.MakeDir(processDir)
	if err != nil {
		return err
	}

	{
		configPath := path.Join(processDir, "conf.yaml")
		conf, err = config.InitConfig(processDir, configPath)
		if err != nil {
			return err
		}

		err = filepath.MakeDir(conf.InstallPath)
		if err != nil {
			return err
		}

		err = filepath.MakeDir(conf.CachePath)
		if err != nil {
			return err
		}
	}

	linkPath = path.Join(processDir, "bin")

	{
		//检查环境变量
		initEnvLinkPath()
		//读取本地安装版本
		readLocalInstallVersion()
		//读取本地缓存列表
		readLocalCacheVersion()
		//读取当前使用的版本
		readCurrentUseVersion()
	}

	app := cli.App{
		Name:        pName,
		HelpName:    "",
		Usage:       "manage go version",
		UsageText:   "",
		ArgsUsage:   "",
		Version:     "0.0.2",
		Description: "a go version manager.\n" + printEnv(),
		Commands: []*cli.Command{
			listCommand(),
			installCommand(),
			useCommand(),
			cacheCommand(),
			uninstallCommand(),
			unuseCommand(),
		},
		BashComplete:           cli.DefaultAppComplete,
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(os.Args)
}

func printEnv() string {

	if !strings.Contains(os.Getenv("PATH"), linkPath) {
		return fmt.Sprintf("\nplease set environment：\033[0;31m%s\033[0;m", linkPath)
	}

	return "\nenvironment set success."
}

func isInstall(info models.Version) bool {
	for _, vInfo := range localInstallVersion {
		if vInfo.Compare(info) == 0 {
			return true
		}
	}
	return false
}

func readLocalCacheVersion() {
	cacheJsonPath := path.Join(conf.CachePath, "version.json")
	buf, err := ioutil.ReadFile(cacheJsonPath)
	if err != nil {
		return
	}
	localCacheVersions = make([]*models.Version, 0)
	json.Unmarshal(buf, &localCacheVersions)

	models.SortV(localCacheVersions).Reverse()
	return
}

func saveLocalCacheVersion() {
	if len(localCacheVersions) == 0 {
		return
	}
	cacheJsonPath := path.Join(conf.CachePath, "version.json")
	file, err := os.OpenFile(cacheJsonPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(localCacheVersions)
}

func readLocalInstallVersion() {
	fileInfoList, err := ioutil.ReadDir(conf.InstallPath)
	if err != nil {
		return
	}
	localInstallVersion = make([]*models.Version, 0)
	for _, info := range fileInfoList {
		if !info.IsDir() {
			continue
		}

		vInfo := models.NewVInfo(info.Name())
		if vInfo.Valid() {
			localInstallVersion = append(localInstallVersion, vInfo)
		}
	}

	models.SortV(localInstallVersion).Reverse()
	return
}

func printError(msg string) {
	fmt.Println("\033[0;31m" + msg + "\033[0;m")
}

func printInfo(msg string) {
	fmt.Println("\033[1;32m" + msg + "\033[0;m")
}

func isInLocalCache(version string) bool {

	for _, cacheVersion := range localCacheVersions {
		if version == cacheVersion.String() {
			return true
		}
	}

	return false
}

func isInInstall(version string) bool {

	for _, cacheVersion := range localInstallVersion {
		if version == cacheVersion.String() {
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
	to = strings.TrimPrefix(to, conf.InstallPath)

	to = strings.ReplaceAll(to, "/", "")
	to = strings.ReplaceAll(to, "\\", "")
	to = strings.TrimSuffix(to, "gobin")
	currentUse.Parse(to)
}

func trimVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "go")
	return version
}

func getCmdLine(cmd ...string) string {
	newCmd := append([]string{pName}, cmd...)
	return strings.Join(newCmd, " ")
}

func printCmdLine(cmd ...string) {
	fmt.Println(getCmdLine(cmd...))
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

func initEnvLinkPath() {
	if strings.Contains(os.Getenv("PATH"), linkPath) {
		return
	}
	if isWin {
		return
	}

	var env string
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") {
		newEnv := path.Join(homeDir, ".bashrc")
		if filepath.FileIsExisted(newEnv) {
			env = newEnv
		} else {
			newEnv = path.Join(homeDir, ".bash_profile")
			if filepath.FileIsExisted(newEnv) {
				env = newEnv
			}
		}
	} else if strings.Contains(shell, "zsh") {
		newEnv := path.Join(homeDir, ".zshrc")
		if filepath.FileIsExisted(newEnv) {
			env = newEnv
		}
	}

	if env == "" {
		for _, s := range []string{".profile", ".bashrc", ".bash_profile", ".zshrc"} {
			newEnv := path.Join(homeDir, s)
			if filepath.FileIsExisted(newEnv) {
				env = newEnv
				break
			}
		}
	}

	if env == "" {
		return
	}

	file, err := os.OpenFile(env, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	if strings.Contains(string(buf), linkPath) {
		return
	}

	//showSetEnv = os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+linkPath) != nil
	//fmt.Println(os.Getenv("PATH"))
	file.WriteString("\nexport PATH=$PATH:")
	file.WriteString(linkPath)
	file.WriteString("\n")
	file.Sync()

	printInfo("\n设置环境变量成功，可能需要重新打开控制台或者注销重新登录才能生效\n")

	return
}
