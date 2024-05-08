package stormi

import (
	"bufio"
	"encoding/json"
	"os"
)

type FileConfig struct {
	ConfigId   string
	ConfigKey  string
	ConfigInfo Config
}

func (fc FileConfig) ToString() string {
	bs, _ := json.MarshalIndent(fc, " ", "  ")
	return string(bs)
}

func (fc FileConfig) ToJson() []byte {
	bs, _ := json.Marshal(fc)
	return bs
}

func (c Config) ToConfigString() string {
	fc := FileConfig{}
	fc.ConfigInfo = c
	fc.ConfigKey = c.Addr + "@" + c.UUID
	fc.ConfigId = c.UUID
	return string(fc.ToJson())
}

func (c Config) ToFileConfig() FileConfig {
	fc := FileConfig{}
	fc.ConfigInfo = c
	fc.ConfigKey = c.Addr + "@" + c.UUID
	fc.ConfigId = c.UUID
	return fc
}

var filename string

func WriteToConfigFile(cs []Config) {
	FileProxy.WriteToFile(filename, nil)
	for _, c := range cs {
		FileProxy.AppendToFile(filename, c.ToConfigString())
	}
}

func AppendToConfigFile(c Config) {
	FileProxy.AppendToFile(filename, c.ToConfigString())
}

func DecoderConfigFile() map[string][]Config {
	file, err := os.Open(filename)
	if err != nil {
		StormiFmtPrintln(magenta, "打开文件时出错: "+err.Error())
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var configMap = map[string][]Config{}
	for scanner.Scan() {
		var fc FileConfig
		if err := json.Unmarshal(scanner.Bytes(), &fc); err != nil {
			StormiFmtPrintln(magenta, "解析 JSON 数据时出错: "+err.Error())
			continue
		}
		configMap[fc.ConfigInfo.Name] = append(configMap[fc.ConfigInfo.Name], fc.ConfigInfo)
	}

	if err := scanner.Err(); err != nil {
		StormiFmtPrintln(magenta, "扫描文件时出错: "+err.Error())
		return nil
	}
	return configMap
}
