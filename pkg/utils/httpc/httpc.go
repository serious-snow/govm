package httpc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"

	"github.com/serious-snow/govm/pkg/utils/path"
)

var client = &http.Client{
	Timeout: time.Minute * 15,
}

func Get(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http请求错误，url: %s，状态码: %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func Download(url, dir, fileName, sha256v string) (returnErr error) {
	if returnErr = path.MakeDir(dir); returnErr != nil {
		return
	}

	newFileName := filepath.Join(dir, fileName)
	tempFileName := newFileName + ".temp"

	file, err := os.Create(tempFileName)
	if err != nil {
		returnErr = err
		return
	}

	var rename bool
	defer func() {
		_ = file.Close()
		if rename {
			returnErr = os.Rename(tempFileName, newFileName)
		}
	}()

	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		returnErr = err
		return
	}

	if resp.StatusCode != http.StatusOK {
		returnErr = fmt.Errorf("下载失败，错误的状态码: %d", resp.StatusCode)
		return
	}

	// 进度条
	bar := pb.Full.Start64(resp.ContentLength)
	tmpl := `{{ bar . "[" "=" ">" " " "]"}} {{counters .}} {{speed . }} {{percent .}}`
	bar.SetTemplateString(tmpl)
	bar.Set(pb.Bytes, true)
	bar.Set(pb.SIBytesPrefix, true)
	defer bar.Finish()

	sha := sha256.New()
	mWriter := io.MultiWriter(file, sha)

	_, err = io.Copy(mWriter, bar.NewProxyReader(resp.Body))
	if err != nil {
		returnErr = err
		return
	}

	dSha256 := hex.EncodeToString(sha.Sum(nil))
	if len(sha256v) != 0 && dSha256 != sha256v {
		_ = os.Remove(tempFileName)
		returnErr = fmt.Errorf("sha256校验不通过，需要: %s, 实际: %s", sha256v, dSha256)
		return
	}
	rename = true
	return
}
