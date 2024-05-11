package stormi

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type fileOpt struct {
}

var FileOpt fileOpt

func (fileOpt) WriteToFile(filename string, ss []string) {
	FileOpt.CreateFileNX(filename)
	for _, s := range ss {
		FileOpt.AppendToFile(filename, s)
	}
}

func (fileOpt) TruncateFile(filename string) {
	FileOpt.CreateFileNX(filename)
	file, _ := os.OpenFile(filename, os.O_RDWR, 0644)

	defer file.Close()

	file.Truncate(0)
}

func (fileOpt) CreateFileNX(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return
		}
		defer file.Close()
	}
}

func (f fileOpt) AppendToFile(filename string, s string) {
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

func (fileOpt) GetMaxConfigFileName(path string) string {
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

func (fileOpt) GetAvailableConfigFileName(path string) string {
	name := FileOpt.GetMaxConfigFileName(path)
	n, err := strconv.Atoi(name)
	if err != nil {
		return name
	}
	return strconv.Itoa(n + 1)
}

func (fileOpt) copyFile(src, dst string) error {
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

func (f fileOpt) copyAllFiles(srcDir, dstDir string) error {
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

func (f fileOpt) Mkdir(path string) {
	os.MkdirAll(path, 0755)
}

func (f fileOpt) IsExistFile(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	return true
}

func isFirstCharUpperCaseLetter(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstRune, _ := utf8.DecodeRuneInString(s)
	return unicode.IsLetter(firstRune) && unicode.IsUpper(firstRune)
}
func pathChange(path string) string {
	os := runtime.GOOS
	if os == "windows" {
		return strings.ReplaceAll(path, "/", "\\")
	} else {
		return strings.ReplaceAll(path, "\\", "/")
	}
}
func (f fileOpt) createCoprotocol(name string) {
	path := f.GoModDir() + "/" + coprotocol + "/" + name
	f.Mkdir(path)
	filename := path + "/" + name + ".go"
	if f.IsExistFile(filename) {
		StormiFmtPrintln(magenta, noredis, "已存在文件:", pathChange(filename))
		return
	}
	f.CreateFileNX(filename)
	f.AppendToFile(filename, "package "+name)
	f.AppendToFile(filename, "\n")
	f.AppendToFile(filename, "//方法")
	f.AppendToFile(filename, "const (")
	f.AppendToFile(filename, "	Func1 = iota")
	f.AppendToFile(filename, "	Func2")
	f.AppendToFile(filename, ")")
	f.AppendToFile(filename, "\n")
	f.AppendToFile(filename, "//数据传输结构体")
	f.AppendToFile(filename, "type "+name+"Dto struct {")
	f.AppendToFile(filename, "	Code    int")
	f.AppendToFile(filename, "	Message string")
	f.AppendToFile(filename, "	Data    string")
	f.AppendToFile(filename, "}")
}

func (f fileOpt) decodeCoprotocol(name string) string {
	path := f.GoModDir() + "/" + coprotocol + "/" + name
	f.Mkdir(path)
	filename := path + "/" + name + ".go"
	if !f.IsExistFile(filename) {
		StormiFmtPrintln(magenta, noredis, "不存在文件:", pathChange(filename))
		return ""
	}
	return f.ReadFile(filename)
}

func (f fileOpt) ReadFile(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(content)
}

func (f fileOpt) GoModDir() string {
	cwd, _ := os.Getwd()
	dir := cwd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		_, err := os.Stat(goModPath)
		if err == nil {
			return dir
		} else if !os.IsNotExist(err) {
			return cwd
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}
