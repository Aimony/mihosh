package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select <group> <proxy>",
	Short: "切换节点",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		if err := proxySvc.SelectProxy(args[0], args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "切换节点失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 已将策略组 '%s' 切换到节点 '%s'\n", args[0], args[1])
	},
}
