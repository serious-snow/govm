package path

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func FileIsExisted(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func PathIsExisted(name string) bool {
	if info, err := os.Stat(name); err == nil {
		return info.IsDir()
	}
	return false
}

// MakeDir 创建文件夹
func MakeDir(dir string) error {
	if !PathIsExisted(dir) {
		return os.MkdirAll(dir, 0777)
	}
	return nil
}

func Decompress(from, to string) error {
	switch filepath.Ext(from) {
	case ".zip":
		return UnZip(from, to)
	default:
		return DecompressTar(from, to)
	}
}

func DecompressTar(from, to string) error {
	fr, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fr.Close()

	// gzip read
	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	// tar read
	tr := tar.NewReader(gr)
	// 读取文件
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if h.FileInfo().IsDir() {
			err = MakeDir(filepath.Join(to, h.Name))
			if err != nil {
				return err
			}
			continue
		}
		// 打开文件
		fw, err := createFile(filepath.Join(to, h.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		//fw.Chmod(0755)
		// 写文件
		_, err = io.Copy(fw, tr)

		if err != nil {
			fw.Close()
			return err
		}
		//
		fw.Close()

	}
	return nil
}

func UnZip(from, to string) error {
	zr, err := zip.OpenReader(from)
	if err != nil {
		return err
	}

	defer zr.Close()

	// 读取文件
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			err = MakeDir(filepath.Join(to, file.Name))
			if err != nil {
				return err
			}
			continue
		}

		if err = copyFile(file, filepath.Join(to, file.Name)); err != nil {
			return err
		}

	}
	return nil
}

func copyFile(from *zip.File, to string) error {
	inFile, err := from.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()
	// 打开文件
	fw, err := createFile(to, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fw.Close()
	_, err = io.Copy(fw, inFile)
	if err != nil {
		return err
	}
	return nil
}

func createFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	dir, _ := filepath.Split(name)

	err := MakeDir(dir)

	if err != nil {
		return nil, err
	}
	return os.OpenFile(name, flag, perm)
}
