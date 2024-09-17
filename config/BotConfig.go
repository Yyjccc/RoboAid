package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var (
	BotConfig *Config
)

type Config struct {
	GroupID    string `yaml:"group_id"`
	AppId      string `yaml:"app_id"`
	SecretKey  string `yaml:"secret_key"`
	ServerPort int    `yaml:"webhook_port"`
}

func init() {
	// 打开 YAML 文件
	file, err := os.Open("bot.yaml")
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	// 解析 YAML 文件
	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("error decoding YAML: %v", err)
	}
	BotConfig = &config
}
