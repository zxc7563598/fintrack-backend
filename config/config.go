package config

import (
	"embed"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	SqlitePath string `yaml:"sqlite_path"`
}

type JWTConfig struct {
	Secret          string        `yaml:"secret"`
	AccessTokenExp  time.Duration `yaml:"access_token_exp"`
	RefreshTokenExp time.Duration `yaml:"refresh_token_exp"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
}

var Cfg Config
var ConfigFS embed.FS

func InitConfig() {
	var file []byte
	var err error

	// 尝试从嵌入的文件系统读取
	if _, openErr := ConfigFS.Open("config.yaml"); openErr == nil {
		file, err = ConfigFS.ReadFile("config.yaml")
		if err != nil {
			log.Fatalf("failed to read embedded config.yaml: %v", err)
		}
	} else {
		// 回退到文件系统读取（开发环境）
		file, err = os.ReadFile("config.yaml")
		if err != nil {
			log.Fatalf("failed to read config.yaml: %v", err)
		}
	}

	if err := yaml.Unmarshal(file, &Cfg); err != nil {
		log.Fatalf("failed to parse config.yaml: %v", err)
	}

	log.Println("✅ config.yaml loaded successfully")
}

func SetConfigFS(fs embed.FS) {
	ConfigFS = fs
}
