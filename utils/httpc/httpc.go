package httpc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/qianlnk/pgbar"
	"govm/utils/filepath"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
)

var (
	client = &http.Client{
		Timeout: time.Minute * 15,
	}
)

func Get(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error http statusCode: %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func Download(url, dir, fileName, sha256v string) error {
	if err := filepath.MakeDir(dir); err != nil {
		return err
	}

	newFileName := path.Join(dir, fileName)
	if filepath.FileIsExisted(newFileName) {

		if sha256v == "" {
			fmt.Println("检测到本地缓存文件存在，且忽略校验")
			return nil
		}

		//校验sha256
		fmt.Println("检测到本地缓存文件存在，开始校验")

		if checkSha256(newFileName, sha256v) {
			fmt.Println("sha256校验通过，无需重新下载")
			return nil
		}
		os.Remove(newFileName)
		fmt.Println("sha256校验不通过，准备重新下载文件")
	}
	tempFileName := newFileName + ".temp"

	file, err := os.Create(tempFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，错误的状态码: %d", resp.StatusCode)
	}
	downloader := NewDownloader(resp)
	sha := sha256.New()
	mWriter := io.MultiWriter(file, sha)

	_, err = io.Copy(mWriter, downloader)

	if err != nil {
		return err
	}

	dSha256 := hex.EncodeToString(sha.Sum(nil))
	if sha256v != "" && dSha256 != sha256v {
		os.Remove(tempFileName)
		return fmt.Errorf("sha256校验不通过，需要: %s,实际: %s", sha256v, dSha256)
	}

	return os.Rename(tempFileName, newFileName)
}

type Downloader struct {
	io.Reader
	bar *pgbar.Bar
}

func NewDownloader(resp *http.Response) *Downloader {
	nb := pgbar.NewBar(0, "下载进度", int(resp.ContentLength))
	if resp.ContentLength > 10*1024 {
		nb.SetUnit("B", "kb", 1024*1024)
	}

	if resp.ContentLength > 10*1024*1024 {
		nb.SetUnit("B", "MB", 1024*1024)
	}
	return &Downloader{
		Reader: resp.Body,
		bar:    nb,
	}
}

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)

	d.bar.Add(n)

	return n, err
}

func checkSha256(fileName, sha256v string) bool {
	sha := sha256.New()
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		return false
	}
	_, err = io.Copy(sha, f)
	if err != nil {
		return false
	}

	dSha256 := hex.EncodeToString(sha.Sum(nil))
	if dSha256 != sha256v {
		return false
	}
	return true
}
