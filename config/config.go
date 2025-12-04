package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIAddress string `mapstructure:"api_address"`
	Secret     string `mapstructure:"secret"`
	TestURL    string `mapstructure:"test_url"`
	Timeout    int    `mapstructure:"timeout"`
}

var defaultConfig = Config{
	APIAddress: "http://127.0.0.1:9090",
	Secret:     "",
	TestURL:    "http://www.gstatic.com/generate_204",
	Timeout:    5000,
}

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
		if err := Save(&defaultConfig); err != nil {
			return nil, err
		}
		return &defaultConfig, nil
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

// Save 保存配置文件
func Save(cfg *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")

	viper.Set("api_address", cfg.APIAddress)
	viper.Set("secret", cfg.Secret)
	viper.Set("test_url", cfg.TestURL)
	viper.Set("timeout", cfg.Timeout)

	return viper.WriteConfigAs(configFile)
}

// Init 初始化配置（交互式）
func Init() error {
	fmt.Println("欢迎使用 Mihomo CLI!")
	fmt.Println("首次使用，需要进行配置初始化")
	fmt.Println()

	cfg := defaultConfig

	fmt.Printf("请输入 Mihomo API 地址 [默认: %s]: ", defaultConfig.APIAddress)
	var input string
	fmt.Scanln(&input)
	if input != "" {
		cfg.APIAddress = input
	}

	fmt.Printf("请输入 API 密钥 (Secret) [可选]: ")
	fmt.Scanln(&input)
	if input != "" {
		cfg.Secret = input
	}

	fmt.Printf("请输入测速 URL [默认: %s]: ", defaultConfig.TestURL)
	fmt.Scanln(&input)
	if input != "" {
		cfg.TestURL = input
	}

	if err := Save(&cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("✓ 配置初始化完成!")
	return nil
}

// Set 设置单个配置项
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "api_address", "api-address":
		cfg.APIAddress = value
	case "secret":
		cfg.Secret = value
	case "test_url", "test-url":
		cfg.TestURL = value
	case "timeout":
		var timeout int
		if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil {
			return fmt.Errorf("timeout 必须是数字: %v", err)
		}
		cfg.Timeout = timeout
	default:
		return fmt.Errorf("未知的配置项: %s (可用: api_address, secret, test_url, timeout)", key)
	}

	return Save(cfg)
}
