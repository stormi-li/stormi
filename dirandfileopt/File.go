package dirandfileopt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func ReadConfigFile(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("读取文件失败：", err)
		return nil
	}
	re := regexp.MustCompile(`<([^>]*)>`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	var matchesContent []string
	for _, match := range matches {
		if len(match) > 1 {
			matchesContent = append(matchesContent, match[1])
		}
	}
	return matchesContent
}

func WriteToConfigFile(nodes []string, path string) error {
	var formattedNodes []string
	for _, node := range nodes {
		formattedNodes = append(formattedNodes, fmt.Sprintf("<%s>", node))
	}
	content := strings.Join(formattedNodes, "\n")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败：%v", err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入文件失败：%v", err)
	}
	return nil
}

func AppendToYaml(filename string, nodes []string) error {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > 0 {
		lines = lines[:len(lines)-1]
	}
	lines = append(lines, "    clusternodes: "+strings.Join(nodes, " "))
	file.Truncate(0)
	file.Seek(0, 0)
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func AppendToConfigFile(filename string, s string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("打开文件时出错:", err)
		return
	}
	defer file.Close()
	_, err = io.WriteString(file, s)
	if err != nil {
		fmt.Println("写入文件时出错:", err)
		return
	}
}

func ExistFile(filename string) bool {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else if err != nil {
		return true
	} else {
		return true
	}
}
