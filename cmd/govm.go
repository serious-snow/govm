package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/v66/github"

	"github.com/serious-snow/govm/pkg/utils/httpc"
	"github.com/serious-snow/govm/pkg/utils/path"
	"github.com/serious-snow/govm/pkg/version"
)

var gitClient = github.NewClient(nil)

const (
	GitUser = "serious-snow"
	GitRepo = "govm"
)

func checkGovmUpdate(ctx context.Context) {
	if Version == "dev" {
		return
	}

	Println("正在拉取 govm 最新版本... ")

	spin := spinner.New(spinner.CharSets[14], time.Millisecond*100)
	spin.Start()

	release, _, err := gitClient.Repositories.GetLatestRelease(ctx, GitUser, GitRepo)
	if err != nil {
		spin.Stop()
		Printf("govm 检查更新失败:%s\n", err)
		return
	}
	spin.Stop()

	sys := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	var asset *github.ReleaseAsset
	for _, v := range release.Assets {
		if strings.Contains(v.GetName(), sys) {
			asset = v
			break
		}
	}
	if asset == nil {
		Printf("govm 已是最新版本\n\n")
		return
	}

	remoteVersion.Govm = GovmVersionInfo{
		Version: release.GetTagName(),
		Size:    asset.GetSize(),
	}

	saveLocalRemoteVersion()
	lastVersion := version.New(release.GetTagName())
	currentVersion := version.New(Version)
	if version.Equal(*lastVersion, *currentVersion) {
		Printf("govm 已是最新版本\n\n")
		return
	}

	Printf("govm 发现新版本：%s，升级命令：%s\n\n", release.GetTagName(), getCmdLine("upgrade govm"))
}

func upgradeGOVM(ctx context.Context) {
	if Version == "dev" {
		return
	}

	Println("正在检查 govm 最新版本")

	release, _, err := gitClient.Repositories.GetLatestRelease(ctx, GitUser, GitRepo)
	if err != nil {
		Println("govm 检查更新失败:", err)
		return
	}
	lastVersion := version.New(release.GetTagName())
	currentVersion := version.New(Version)
	if version.Equal(*lastVersion, *currentVersion) {
		Println("govm 已是最新版本")
		return
	}

	Printf("正在升级 govm %s --> %s\n", Version, release.GetTagName())

	sys := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	var asset *github.ReleaseAsset
	for _, v := range release.Assets {
		if strings.Contains(v.GetName(), sys) {
			asset = v
			break
		}
	}
	if asset == nil {
		Println("govm 升级包未找到", sys)
		return
	}

	tempDir, err := os.MkdirTemp("", "govm")
	if err != nil {
		Println("govm 下载失败", err)
		return
	}
	defer os.RemoveAll(tempDir)

	tempFileName := filepath.Join(os.TempDir(), asset.GetName())

	fd, fp := filepath.Split(tempFileName)

	Println("下载:", asset.GetBrowserDownloadURL(), "-->", tempFileName)
	if err := httpc.Download(asset.GetBrowserDownloadURL(), fd, fp, ""); err != nil {
		Println("govm 下载失败：", err)
		return
	}

	if err := path.Decompress(tempFileName, tempDir); err != nil {
		Println("govm 解压失败：", err)
		return
	}

	binFile := "govm"
	if runtime.GOOS == "windows" {
		binFile += ".exe"
	}

	tempFile := os.Args[0] + ".new"
	execFile := filepath.Join(tempDir, binFile)
	err = os.Rename(execFile, tempFile)
	if err != nil {
		Println("govm 升级失败：", err)
		return
	}

	_ = os.Chmod(tempFile, os.ModePerm)
	err = replaceExecutable(tempFile, getExecutable())
	if err != nil {
		Println("govm 升级失败：", err)
		return
	}

	Println("govm 升级成功")
}

func getExecutable() string {
	ex, err := os.Executable()
	if err != nil {
		return os.Args[0]
	}
	ev, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return ex
	}
	return ev
}
