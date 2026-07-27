package util

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	RedisPassword        string        `mapstructure:"REDIS_PASSWORD"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	ServerPort           string        `mapstructure:"SERVER_PORT"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	OSSDomain            string        `mapstructure:"OSS_DOMAIN"`
	MinioEndpoint        string        `mapstructure:"MINIO_ENDPOINT"`
	MinioEndpointLocal   string        `mapstructure:"MINIO_ENDPOINT_LOCAL"`
	MinioAccessKeyID     string        `mapstructure:"MINIO_ACCESS_KEY_ID"`
	MinioSecretAccessKey string        `mapstructure:"MINIO_SECRET_ACCESS_KEY"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal("初始化配置错误", err)
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

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

func (c *Config) IsLoadTest() bool {
	return strings.ToLower(c.Environment) == "loadtest"
}

// InitTestConfig 智能向上寻找并加载配置
func InitTestConfig() Config {
	// 从当前运行的子包目录开始，一层一层往上盲找 app.env
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 6; i++ {
		envPath := filepath.Join(dir, "app.env")
		if _, err := os.Stat(envPath); err == nil { // 找到了
			return InitConfig(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	panic("failed to find app.env in any parent directory during test")
}
