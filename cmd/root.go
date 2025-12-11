package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/internal/ui/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mihosh",
	Short: "Mihosh - Mihomo 终端管理工具",
	Long:  `一个功能完整的 mihomo 终端命令行工具，支持节点切换、测速等操作`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认行为：启动TUI界面
		cfg, err := config.Load()
		if err != nil {
			// 友好的首次使用引导
			configSvc := service.NewConfigService()
			if err := configSvc.InitConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "配置初始化失败: %v\n", err)
				os.Exit(1)
			}

			// 重新加载配置
			cfg, err = config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
				os.Exit(1)
			}
		}

		client := api.NewClient(cfg)
		model := tui.NewModel(client, cfg.TestURL, cfg.Timeout)

		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(testGroupCmd)
	rootCmd.AddCommand(connectionsCmd)
}

// Execute 执行命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
