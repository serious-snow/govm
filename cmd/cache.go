package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/serious-snow/govm/pkg/utils/path"
)

func cacheCommand() *cli.Command {
	return &cli.Command{
		Name:      "cache",
		Aliases:   []string{"c"},
		Usage:     "Cache manager",
		UsageText: getCmdLine("cache", "[dir]", "[clear]"),
		Commands: []*cli.Command{
			{
				Name:  "dir",
				Usage: "Print cache dir",
				Action: func(context *cli.Context) error {
					Println(conf.CachePath)
					return nil
				},
			},
			{
				Name:  "clear",
				Usage: "Clear cache",
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

						switch filepath.Ext(info.Name()) {
						case ".json":
						default:
							_ = os.Remove(filepath.Join(conf.CachePath, info.Name()))
						}

					}
					return nil
				},
			},
			{
				Name:  "size",
				Usage: "Show cache size",
				Action: func(context *cli.Context) error {
					size := int64(0)
					defer func() {
						Println(formatSize(size))
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

						switch filepath.Ext(info.Name()) {
						case ".json":
						default:
							size += info.Size()
						}

					}
					return nil
				},
			},
		},
	}
}
