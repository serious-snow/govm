package cmd

import "github.com/serious-snow/govm/pkg/version"

type (
	// RemoteVersion 版本信息
	RemoteVersion struct {
		GovmVersion string             `json:"govm_version"`
		GoVersions  []*version.Version `json:"go_versions"`
	}
)
