package util

import (
	"log"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Environment  string `mapstructure:"ENVIRONMENT"`
	DBSource     string `mapstructure:"DB_SOURCE"`
	MigrationURL string `mapstructure:"MIGRATION_URL"`
	SERVER_PORT  string `mapstructure:"SERVER_PORT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal("初始化配置错误")
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

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
