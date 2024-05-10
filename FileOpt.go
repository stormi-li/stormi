package stormi

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type FileOpt struct {
}

var FileProxy FileOpt

func (FileOpt) WriteToFile(filename string, ss []string) {
	FileProxy.CreateFileNX(filename)
	// FileProxy.TruncateFile(filename)
	for _, s := range ss {
		FileProxy.AppendToFile(filename, s)
	}
}

func (FileOpt) TruncateFile(filename string) {
	FileProxy.CreateFileNX(filename)
	file, _ := os.OpenFile(filename, os.O_RDWR, 0644)

	defer file.Close()

	file.Truncate(0)
}

func (FileOpt) CreateFileNX(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return
		}
		defer file.Close()
	}
}

func (f FileOpt) AppendToFile(filename string, s string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		StormiFmtPrintln(magenta, "打开文件时出错"+err.Error())
		return
	}
	defer file.Close()
	s = s + "\n"
	if s == "" {
		return
	}
	_, err = io.WriteString(file, s)
	if err != nil {
		StormiFmtPrintln(magenta, "写入文件时出错"+err.Error())
		return
	}
}

func (FileOpt) GetMaxConfigFileName(path string) string {
	dirPath := path

	files, err := os.ReadDir(dirPath)
	if err != nil {
		StormiFmtPrintln(magenta, "读取目录时出错"+err.Error())
		return "1"
	}

	maxFileName := ""
	maxIndex := 0
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			ext := filepath.Ext(fileName)
			name := fileName[:len(fileName)-len(ext)]
			if name != "" {
				index, err := strconv.Atoi(name)
				if err == nil && index > maxIndex {
					maxIndex = index
					maxFileName = name
				}
			}
		}
	}
	if maxFileName == "" {
		return "1"
	}

	n, err := strconv.Atoi(maxFileName)
	if err != nil {
		return maxFileName
	}
	return strconv.Itoa(n + 1)
}

func (FileOpt) GetAvailableConfigFileName(path string) string {
	name := FileProxy.GetMaxConfigFileName(path)
	n, err := strconv.Atoi(name)
	if err != nil {
		return name
	}
	return strconv.Itoa(n + 1)
}

func (FileOpt) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func (f FileOpt) copyAllFiles(srcDir, dstDir string) error {
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != srcDir {
			dstPath := filepath.Join(dstDir, info.Name())
			if info.Mode().IsRegular() {
				err := f.copyFile(path, dstPath)
				if err != nil {
					return err
				}
				StormiFmtPrintln(yellow, noredis, "已安装:", info.Name(), "到", dstDir)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
