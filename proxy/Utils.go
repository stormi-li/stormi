package proxy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/stormi-li/stormi/formatprint"
)

type ticker struct {
	time time.Time
}

type utils struct{}

var Utils utils

func (utils) NewTicker() ticker {
	t := ticker{
		time: time.Now(),
	}
	return t
}

func (tt ticker) Stamp(str string) {
	t := time.Now()
	elapsed := t.Sub(tt.time)
	fmt.Println(str, "consumed", elapsed)
}

func (utils) readConfigFile(path string) []string {
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

func (utils) writeToConfigFile(nodes []string, path string) error {
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

func (utils) appendToYaml(filename string, nodes []string) error {
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

type code struct{}

var Shell code

func appendToConfigFile(s string) {
	filename := currentDir + "/app.config"
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

func sh(cmd string) string {
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return strings.ReplaceAll(strings.ReplaceAll(string(output), "\n", ""), "\r", "")
}

func sha(cmd string) {
	output, _ := exec.Command("bash", "-c", cmd).CombinedOutput()
	fmt.Print(string(output))
}

func (code) StormiVersion() {
	sha("stormi version")
}

func (code) PortStatus(port int) {
	if port == -1 {
		sha("stormi port-list ")
		return
	}
	res := sh("stormi port-list " + strconv.Itoa(port))
	if res == "" {
		fmt.Println("该端口并未被任何进程占用")
	} else {
		fmt.Println(res)
	}
}

func (code) PortProcess(port int) int {
	res := sh("$(stormi-scriptdir.sh)/stormi-port/stormi-getportprocess.sh " + strconv.Itoa(port))
	i, err := strconv.Atoi(res)
	if err != nil {
		return -1
	}
	return i
}

func (code) KillProcess(processid int) {
	sha("kill " + strconv.Itoa(processid))
}

func ExecCommand(cmd string) {
	os := runtime.GOOS
	var out []byte
	var err error
	if os == "windows" {
		cmd = strings.Replace(cmd, "/", "\\", -1)
		out, err = exec.Command("cmd", "/C", cmd).CombinedOutput()
	} else {
		out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	}

	if err != nil {
		formatprint.FormatPrint(formatprint.Red, err.Error())
		return
	}
	formatprint.FormatPrint(formatprint.Green, "脚本执行结果:\n"+string(out))
}
