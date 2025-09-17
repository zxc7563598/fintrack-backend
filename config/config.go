package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	SqlitePath string `yaml:"sqlite_path"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret          string        `yaml:"secret"`
	AccessTokenExp  time.Duration `yaml:"access_token_exp"`
	RefreshTokenExp time.Duration `yaml:"refresh_token_exp"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
}

var Cfg Config

func InitConfig() {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("failed to read config.yaml: %v", err)
	}

	if err := yaml.Unmarshal(file, &Cfg); err != nil {
		log.Fatalf("failed to parse config.yaml: %v", err)
	}

	log.Println("✅ config.yaml loaded successfully")
}
