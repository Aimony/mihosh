package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihomo-cli/internal/app/service"
	"github.com/aimony/mihomo-cli/internal/infrastructure/api"
	"github.com/aimony/mihomo-cli/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test <proxy>",
	Short: "测速单个代理节点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		delay, err := proxySvc.TestProxyDelay(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "测速失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 节点 '%s' 延迟: %dms\n", args[0], delay)
	},
}

var testGroupCmd = &cobra.Command{
	Use:   "test-group <group>",
	Short: "测速策略组内所有节点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		if err := proxySvc.TestGroupDelay(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "批量测速失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 策略组 '%s' 测速完成\n", args[0])
	},
}
