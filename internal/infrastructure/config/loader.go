package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ErrConfigNotFound 配置文件不存在
var ErrConfigNotFound = errors.New("配置文件不存在")

// GetConfigDir 获取配置目录
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".mihosh")
	return configDir, nil
}

// Load 加载配置文件
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// 配置文件不存在时返回错误，由调用方决定是否触发初始化引导
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
