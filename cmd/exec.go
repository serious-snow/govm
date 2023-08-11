package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v3"
)

func execCommand() *cli.Command {
	return &cli.Command{
		Name:      "exec",
		Aliases:   []string{"e"},
		Usage:     "Exec command with the PATH pointing to go version",
		UsageText: getCmdLine("exec", "<version>", "go build main.go"),
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return cli.ShowSubcommandHelp(c)
			}
			version := c.Args().Get(0)
			if !isInInstall(version) {
				printError("该版本未安装，请先安装，执行：")
				printCmdLine("install", version)
				return nil
			}

			goRoot := filepath.Join(conf.InstallPath, version, "go")
			goBin := filepath.Join(goRoot, "bin")
			goToolDir := filepath.Join(goRoot, "/pkg/tools/", runtime.GOOS+"_"+runtime.GOARCH)
			if err := os.Setenv("GOTOOLDIR", goToolDir); err != nil {
				return err
			}

			if err := os.Setenv("GOROOT", goRoot); err != nil {
				return err
			}

			newPath := goBin + string(os.PathListSeparator) + os.Getenv("PATH")
			if err := os.Setenv("PATH", newPath); err != nil {
				return err
			}

			args := c.Args().Slice()
			cmd := exec.Command(args[1], args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}
