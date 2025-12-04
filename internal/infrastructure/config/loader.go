package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// GetConfigDir 获取配置目录
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".mihomo-cli")
	return configDir, nil
}

// Load 加载配置文件
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := Save(&DefaultConfig); err != nil {
			return nil, err
		}
		return &DefaultConfig, nil
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
