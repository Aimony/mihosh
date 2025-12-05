package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置",
	Run: func(cmd *cobra.Command, args []string) {
		configSvc := service.NewConfigService()
		if err := configSvc.InitConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "初始化配置失败: %v\n", err)
			os.Exit(1)
		}
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		configSvc := service.NewConfigService()
		cfg, err := configSvc.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("当前配置:")
		fmt.Printf("  API 地址: %s\n", cfg.APIAddress)
		fmt.Printf("  密钥:     %s\n", utils.MaskSecret(cfg.Secret))
		fmt.Printf("  测速 URL: %s\n", cfg.TestURL)
		fmt.Printf("  超时:     %dms\n", cfg.Timeout)

		// 显示配置文件位置
		configDir, _ := config.GetConfigDir()
		fmt.Printf("\n配置文件位置: %s\\config.yaml\n", configDir)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项",
	Long: `设置配置项

可用的配置项:
  api-address  - Mihomo API 地址 (例如: http://127.0.0.1:9090)
  secret       - API 密钥
  test-url     - 测速 URL (例如: http://www.gstatic.com/generate_204)
  timeout      - 超时时间，单位毫秒 (例如: 5000)

示例:
  mihomo config set api-address http://127.0.0.1:9090
  mihomo config set secret your-secret-here
  mihomo config set test-url http://www.google.com/generate_204
  mihomo config set timeout 3000`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		configSvc := service.NewConfigService()
		if err := configSvc.SetConfigValue(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "设置配置失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 已设置 %s = %s\n", key, value)
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}
