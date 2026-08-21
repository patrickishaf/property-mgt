package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
}

var config Config

func GetConfig() *Config {
	return &config
}

func LoadConfig() *Config {
	filePath := configFilePath()
	if err := godotenv.Load(filePath); err != nil {
		panic(err)
	}
	if err := envconfig.Process("", &config); err != nil {
		panic(err)
	}
	return &config
}

func configFilePath() string {
	_, err := os.Stat(".env")
	if err == nil {
		return ".env"
	}
	if os.IsNotExist(err) {
		_, err := os.Stat("../../.env")
		if err == nil {
			return "../../.env"
		}
		if os.IsNotExist(err) {
			log.Fatal("config file not found")
		}
	}
	log.Fatal("config file not found")
	return ""
}
