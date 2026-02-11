package cmd

import "github.com/serious-snow/govm/pkg/version"

type (
	// RemoteVersion 版本信息
	RemoteVersion struct {
		Govm GovmVersionInfo  `json:"govm"`
		Go   []*GoVersionInfo `json:"go"`
	}
)

type GoVersionInfo struct {
	Filename string          `json:"filename"`
	Version  version.Version `json:"version"`
	Sha256   string          `json:"sha256"`
	Size     int             `json:"size"`
}

type GovmVersionInfo struct {
	Version string `json:"version"`
	Size    int    `json:"size"`
}

type ListGoVersionResponse struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []*struct {
		Filename string `json:"filename"`
		Os       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
		Sha256   string `json:"sha256"`
		Size     int    `json:"size"`
		Kind     string `json:"kind"`
	} `json:"files"`
}
