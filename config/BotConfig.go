package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var (
	BotConfig *Config
	BotPath   string
)

type Config struct {
	GroupID     string      `yaml:"group_id"`
	DbPath      string      `yaml:"db_path"`
	AppId       string      `yaml:"app_id"`
	SecretKey   string      `yaml:"secret_key"`
	ServerPort  int         `yaml:"webhook_port"`
	VerifyToken string      `yaml:"verify_token"`
	Owner       string      `yaml:"owner"`
	Templates   []*Template `yaml:"templates"`
}

func (c *Config) GetTmpl(name string) *Template {
	for _, t := range c.Templates {
		if t.Name == name {
			return t
		}
	}
	return nil
}

type Template struct {
	Name    string `yaml:"name"`
	ID      string `yaml:"id"`
	Version string `yaml:"version"`
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
	BotPath, err = os.Getwd()
}
