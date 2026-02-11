//go:build !windows

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/serious-snow/govm/pkg/utils/path"
)

func SetEnv() {
	if strings.Contains(os.Getenv("PATH"), envPath) {
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

	file, err := os.OpenFile(env, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return
	}

	if strings.Contains(string(buf), envPath) {
		return
	}

	// showSetEnv = os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+linkPath) != nil
	// Println(os.Getenv("PATH"))
	_, _ = file.WriteString("\nexport PATH=$PATH:")
	_, _ = file.WriteString(envPath)
	_, _ = file.WriteString("\n")
	_ = file.Sync()

	printInfo(fmt.Sprintf("\n环境变量设置于 %s\n可能需要重新打开控制台或者注销重新登录才能生效\n", env))
}

func Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func replaceExecutable(currentPath, newVersionPath string) error {
	return os.Rename(currentPath, newVersionPath)
}
