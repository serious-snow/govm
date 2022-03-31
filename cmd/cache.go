package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"govm/utils/path"
	"io/ioutil"
	"os"
	"path/filepath"
)

func cacheCommand() *cli.Command {
	return &cli.Command{
		Name:      "cache",
		Aliases:   []string{"c"},
		Usage:     "cache manager",
		UsageText: getCmdLine("cache", "[dir]", "[clean]"),
		Subcommands: []*cli.Command{
			{
				Name:  "dir",
				Usage: "print cache dir",
				Action: func(context *cli.Context) error {
					fmt.Println(conf.CachePath)
					return nil
				},
			},
			{
				Name:  "clean",
				Usage: "clean cache dir",
				Action: func(context *cli.Context) error {
					if !path.PathIsExisted(conf.CachePath) {
						return nil
					}
					fileInfoList, err := ioutil.ReadDir(conf.CachePath)
					if err != nil {
						return nil
					}
					for _, info := range fileInfoList {

						if info.IsDir() {
							continue
						}

						if info.Name() == "version.json" {
							continue
						}

						os.Remove(filepath.Join(conf.CachePath, info.Name()))
					}
					return nil
				},
			},
			{
				Name:  "size",
				Usage: "show cache size",
				Action: func(context *cli.Context) error {
					size := int64(0)
					defer func() {
						fmt.Println(formatSize(size))
					}()
					if !path.PathIsExisted(conf.CachePath) {
						return nil
					}
					fileInfoList, err := ioutil.ReadDir(conf.CachePath)
					if err != nil {
						return nil
					}
					for _, info := range fileInfoList {

						if info.IsDir() {
							continue
						}

						if info.Name() == "version.json" {
							continue
						}

						size += info.Size()
					}
					return nil
				},
			},
		},
	}
}
