package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihomo-cli/api"
	"github.com/aimony/mihomo-cli/config"
	"github.com/aimony/mihomo-cli/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mihomo",
	Short: "Mihomo CLI - 终端操作 mihomo 代理",
	Long:  `一个功能完整的 mihomo 终端命令行工具，支持节点切换、测速等操作`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认行为：启动TUI界面
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			fmt.Println("请先运行 'mihomo config init' 初始化配置")
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		model := ui.NewModel(client, cfg.TestURL, cfg.Timeout)

		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
			os.Exit(1)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "初始化配置失败: %v\n", err)
			os.Exit(1)
		}
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("当前配置:")
		fmt.Printf("  API 地址: %s\n", cfg.APIAddress)
		fmt.Printf("  密钥: %s\n", maskSecret(cfg.Secret))
		fmt.Printf("  测速 URL: %s\n", cfg.TestURL)
		fmt.Printf("  超时时间: %dms\n", cfg.Timeout)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有策略组和节点",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		groups, err := client.GetGroups()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取策略组失败: %v\n", err)
			os.Exit(1)
		}

		proxies, err := client.GetProxies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取代理信息失败: %v\n", err)
			os.Exit(1)
		}

		for name, group := range groups {
			fmt.Printf("\n策略组: %s [%s]\n", name, group.Type)
			fmt.Printf("当前节点: %s\n", group.Now)
			fmt.Println("可用节点:")
			for _, proxyName := range group.All {
				proxy := proxies[proxyName]
				delay := ""
				if len(proxy.History) > 0 {
					lastDelay := proxy.History[len(proxy.History)-1].Delay
					if lastDelay > 0 {
						delay = fmt.Sprintf(" (%dms)", lastDelay)
					}
				}
				marker := ""
				if proxyName == group.Now {
					marker = " ✓"
				}
				fmt.Printf("  - %s%s%s\n", proxyName, delay, marker)
			}
		}
	},
}

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
		if err := client.SelectProxy(args[0], args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "切换节点失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 已将策略组 '%s' 切换到节点 '%s'\n", args[0], args[1])
	},
}

var testCmd = &cobra.Command{
	Use:   "test <proxy>",
	Short: "测速指定节点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		delay, err := client.TestProxyDelay(args[0], cfg.TestURL, cfg.Timeout)
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
		if err := client.TestGroupDelay(args[0], cfg.TestURL, cfg.Timeout); err != nil {
			fmt.Fprintf(os.Stderr, "测速失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 策略组 '%s' 测速完成\n", args[0])
	},
}

var connectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "查看当前连接",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		conns, err := client.GetConnections()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取连接失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("总连接数: %d\n", len(conns.Connections))
		fmt.Printf("总上传: %s\n", formatBytes(conns.UploadTotal))
		fmt.Printf("总下载: %s\n", formatBytes(conns.DownloadTotal))
		fmt.Println("\n活跃连接:")
		for _, conn := range conns.Connections {
			fmt.Printf("  %s:%s -> %s:%s [%s]\n",
				conn.Metadata.SourceIP,
				conn.Metadata.SourcePort,
				conn.Metadata.DestinationIP,
				conn.Metadata.DestinationPort,
				conn.Chains[len(conn.Chains)-1],
			)
		}
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	
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

// 辅助函数
func maskSecret(secret string) string {
	if secret == "" {
		return "(未设置)"
	}
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:2] + "****" + secret[len(secret)-2:]
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
