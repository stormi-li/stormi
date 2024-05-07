package dirandfileopt

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stormi-li/stormi/formatprint"
)

func CreateDirInCurrentDir(name string) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// 定义你想要创建的目录名
	newDirName := name

	// 构建完整的目录路径（如果只是在当前目录下创建，则使用当前目录路径+目录名）
	newDirFullPath := filepath.Join(currentDir, newDirName)

	// 使用MkdirAll创建目录，它会递归地创建所有必要的上级目录
	err = os.MkdirAll(newDirFullPath, 0755) // 0755 是目录权限（rwxr-xr-x）
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	formatprint.FormatPrint(formatprint.Blue, "目录创建成功:"+newDirFullPath)
}

func CreateDir(dir string) {
	err := os.MkdirAll(dir, 0755) // 0755 是目录权限（rwxr-xr-x）
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	formatprint.FormatPrint(formatprint.Blue, "目录创建成功:"+dir)
}

func ExistDir(path string) bool {
	dirPath := path

	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
