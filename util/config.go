package util

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

const (
	MinRegion  = 0
	MaxRegion  = 13
	DateFormat = "2006-01-02"
)

var (
	MinDate     = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	MaxDate     = time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	ExpiredDate = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
)

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	DBSource    string `mapstructure:"DB_SOURCE"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("初始化配置错误")
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func InitConfig(path string) Config {
	config, err := LoadConfig(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if config.Environment == "development" {
		log.Printf("当前处于开发模式下")
	}

	return config
}
