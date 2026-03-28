package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
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

		groups, orderedNames, err := proxySvc.GetGroups()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取策略组失败: %v\n", err)
			os.Exit(1)
		}

		proxiesMap, err := proxySvc.GetProxies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取代理失败: %v\n", err)
			os.Exit(1)
		}

		renderGroupList(os.Stdout, groups, orderedNames, proxiesMap)
	},
}

func renderGroupList(w io.Writer, groups map[string]model.Group, orderedNames []string, proxiesMap map[string]model.Proxy) {
	fmt.Fprintln(w, "策略组列表:")
	for _, groupName := range resolveGroupOrder(groups, orderedNames) {
		group := groups[groupName]
		fmt.Fprintf(w, "\n[%s] %s (当前: %s)\n", group.Type, group.Name, group.Now)
		for _, proxyName := range group.All {
			proxy, ok := proxiesMap[proxyName]
			delay := ""
			if ok && len(proxy.History) > 0 {
				lastDelay := proxy.History[len(proxy.History)-1].Delay
				if lastDelay > 0 {
					delay = fmt.Sprintf(" (%dms)", lastDelay)
				}
			}
			marker := ""
			if proxyName == group.Now {
				marker = " ✓"
			}
			fmt.Fprintf(w, "  - %s%s%s\n", proxyName, delay, marker)
		}
	}
}

func resolveGroupOrder(groups map[string]model.Group, orderedNames []string) []string {
	if len(groups) == 0 {
		return nil
	}

	if len(orderedNames) == 0 {
		names := make([]string, 0, len(groups))
		for name := range groups {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}

	names := make([]string, 0, len(groups))
	seen := make(map[string]struct{}, len(groups))

	for _, name := range orderedNames {
		if _, ok := groups[name]; ok {
			names = append(names, name)
			seen[name] = struct{}{}
		}
	}

	if len(names) == len(groups) {
		return names
	}

	extras := make([]string, 0, len(groups)-len(names))
	for name := range groups {
		if _, ok := seen[name]; !ok {
			extras = append(extras, name)
		}
	}
	sort.Strings(extras)

	return append(names, extras...)
}
