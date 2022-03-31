package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/config"
	"govm/models"
	"govm/utils/path"
	"io/ioutil"
	"os"
	"path/filepath"
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
	linkPath string //软连接路径
	envPath  string //环境变量路径
)

func Run() error {
	pName = filepath.Base(os.Args[0])
	isWin = runtime.GOOS == "windows"

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
		//检查环境变量
		initEnvPath()
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
		Version:     "0.0.3",
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
	cacheJsonPath := filepath.Join(conf.CachePath, "version.json")
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
	cacheJsonPath := filepath.Join(conf.CachePath, "version.json")
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
	to = strings.TrimSuffix(to, "go")
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

func initEnvPath() {
	if strings.Contains(os.Getenv("PATH"), envPath) {
		return
	}
	if isWin {
		return
	}

	var env string
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") {
		newEnv := filepath.Join(homeDir, ".bashrc")
		if path.FileIsExisted(newEnv) {
			env = newEnv
		} else {
			newEnv = filepath.Join(homeDir, ".bash_profile")
			if path.FileIsExisted(newEnv) {
				env = newEnv
			}
		}
	} else if strings.Contains(shell, "zsh") {
		newEnv := filepath.Join(homeDir, ".zshrc")
		if path.FileIsExisted(newEnv) {
			env = newEnv
		}
	}

	if env == "" {
		for _, s := range []string{".profile", ".bashrc", ".bash_profile", ".zshrc"} {
			newEnv := filepath.Join(homeDir, s)
			if path.FileIsExisted(newEnv) {
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

	if strings.Contains(string(buf), envPath) {
		return
	}

	//showSetEnv = os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+linkPath) != nil
	//fmt.Println(os.Getenv("PATH"))
	file.WriteString("\nexport PATH=$PATH:")
	file.WriteString(envPath)
	file.WriteString("\n")
	file.Sync()

	printInfo("\n设置环境变量成功，可能需要重新打开控制台或者注销重新登录才能生效\n")

	return
}

func getDownloadFilename(version string) string {
	suffix := "tar.gz"
	if isWin {
		suffix = "zip"
	}

	return fmt.Sprintf("go%s.%s-%s.%s", version, runtime.GOOS, runtime.GOARCH, suffix)
}
