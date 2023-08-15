//go:build windows

package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows/registry"

	"github.com/serious-snow/govm/pkg/utils/path"
)

func SetEnv() {

	if strings.Contains(os.Getenv("PATH"), envPath) {
		return
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		ErrorLn("无法打开注册表键:", err)
		return
	}
	defer key.Close()

	oldPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		ErrorLn("无法获取环境变量:", err)
		return
	}

	if strings.Contains(oldPath, envPath) {
		return
	}

	newPath := oldPath
	if len(newPath) == 0 {
		newPath = envPath
	} else {
		newPath = strings.Join([]string{newPath, envPath}, string(os.PathListSeparator))
	}

	cmd := exec.Command("setx", "PATH", newPath)

	// 设置输出和错误输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行命令
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to run setx command:", err)
		return
	}

	printInfo("\n设置环境变量成功，可能需要重新打开控制台或者注销重新登录才能生效\n")
}

func Symlink(oldname, newname string) error {
	err := os.Symlink(oldname, newname)
	if err == nil {
		return nil
	}

	return UacSymlink(oldname, newname)
}

func UacSymlink(oldname, newname string) error {

	//c := strings.Join([]string{tempF.Name(), "symlink", oldname, newname}, " ")
	//Copy
	//mklink /D "path_to_link_directory" "path_to_target_directory"
	c := strings.Join([]string{"mklink", "/D", newname, oldname}, " ")
	cmd := exec.Command("cmd.exe", "/C", c)
	//cmd = exec.Command("runas", "/user:Administrator", c)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := elevate(cmd); err != nil {
		return err
	}

	// Must
	time.Sleep(time.Millisecond * 200)

	if !path.PathIsExisted(newname) {
		return errors.New("没有足够的权限，请使用管理员重试")
	}

	return nil
}

func elevate(cmd *exec.Cmd) error {
	if err := ole.CoInitialize(0); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	shell, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return err
	}
	defer shell.Release()

	shellDispatch, err := shell.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer shellDispatch.Release()

	verb := "runas"
	filePath, err := exec.LookPath(cmd.Path)
	if err != nil {
		return err
	}

	params := strings.Join(cmd.Args, " ")

	_, err = oleutil.CallMethod(shellDispatch, "ShellExecute", filePath, params, "", verb)
	if err != nil {
		return err
	}

	return nil
}
