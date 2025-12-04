package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihomo-cli/internal/app/service"
	"github.com/aimony/mihomo-cli/internal/infrastructure/api"
	"github.com/aimony/mihomo-cli/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

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
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		groups, err := proxySvc.GetGroups()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取策略组失败: %v\n", err)
			os.Exit(1)
		}

		proxiesMap, err := proxySvc.GetProxies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取代理失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("策略组列表:")
		for _, group := range groups {
			fmt.Printf("\n[%s] %s (当前: %s)\n", group.Type, group.Name, group.Now)
			for _, proxyName := range group.All {
				proxy := proxiesMap[proxyName]
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
