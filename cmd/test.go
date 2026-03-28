package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

type testAction string

const (
	actionCurrent testAction = "current"
	actionNode    testAction = "node"
	actionGroup   testAction = "group"
)

var testCmd = &cobra.Command{
	Use:   "test [node <节点名> | group <策略组名>]",
	Short: "测试节点功能",
	Args: func(cmd *cobra.Command, args []string) error {
		_, _, err := resolveTestAction(args)
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		action, target, err := resolveTestAction(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if err := runTestAction(proxySvc, cfg.ProxyAddress, action, target); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

var testGroupCmd = &cobra.Command{
	Use:    "test-group <group>",
	Short:  "[兼容] 测试指定策略组里的所有节点",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		if err := runTestAction(proxySvc, cfg.ProxyAddress, actionGroup, args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

func resolveTestAction(args []string) (testAction, string, error) {
	if len(args) == 0 {
		return actionCurrent, "", nil
	}

	if len(args) == 2 {
		switch args[0] {
		case string(actionNode):
			return actionNode, args[1], nil
		case string(actionGroup):
			return actionGroup, args[1], nil
		}
	}

	return "", "", fmt.Errorf("参数格式错误。请使用：mihosh test | mihosh test node <节点名> | mihosh test group <策略组名>")
}

func runTestAction(proxySvc *service.ProxyService, proxyAddress string, action testAction, target string) error {
	switch action {
	case actionCurrent:
		node, found, err := currentSelectedNode(proxySvc)
		if err != nil {
			return fmt.Errorf("获取当前选中节点失败: %w", err)
		}
		if !found {
			fmt.Println("未检测到当前选中的节点")
			return nil
		}

		// 先验证当前选中节点可测速，再保留原本的链路/IP信息输出格式。
		_, err = proxySvc.TestProxyDelay(node)
		if err != nil {
			return fmt.Errorf("测速失败: %w", err)
		}

		chain, err := proxySvc.GetNodeChain()
		if err != nil {
			return fmt.Errorf("获取节点链路失败: %w", err)
		}

		ipInfo, err := proxySvc.GetIPInfo(proxyAddress)
		if err != nil {
			return fmt.Errorf("获取 IP 信息失败: %w", err)
		}

		const contentWidth = 50
		labels := []string{"节点链路", "节点 IP", "国家/地区", "城市", "ASN", "组织"}
		labelWidth := maxDisplayWidth(labels)
		if labelWidth < 8 {
			labelWidth = 8
		}

		fmt.Println("┌" + strings.Repeat("─", contentWidth+2) + "┐")
		fmt.Println(boxLine("节点链路", strings.Join(chain, " -> "), labelWidth, contentWidth))
		fmt.Println(boxLine("节点 IP", ipInfo.IP, labelWidth, contentWidth))
		fmt.Println(boxLine("国家/地区", fmt.Sprintf("%s (%s)", ipInfo.Country, ipInfo.CountryCode), labelWidth, contentWidth))
		fmt.Println(boxLine("城市", ipInfo.City, labelWidth, contentWidth))
		fmt.Println(boxLine("ASN", ipInfo.AS, labelWidth, contentWidth))
		fmt.Println(boxLine("组织", ipInfo.Org, labelWidth, contentWidth))
		fmt.Println("└" + strings.Repeat("─", contentWidth+2) + "┘")
		return nil

	case actionNode:
		delay, err := proxySvc.TestProxyDelay(target)
		if err != nil {
			return fmt.Errorf("测速失败: %w", err)
		}
		fmt.Printf("✓ 节点 '%s' 延迟: %dms\n", target, delay)
		return nil

	case actionGroup:
		if err := proxySvc.TestGroupDelay(target); err != nil {
			return fmt.Errorf("批量测速失败: %w", err)
		}
		fmt.Printf("✓ 策略组 '%s' 测速完成\n", target)
		return nil
	}

	return fmt.Errorf("不支持的测试动作: %s", action)
}

func currentSelectedNode(proxySvc *service.ProxyService) (string, bool, error) {
	proxies, err := proxySvc.GetProxies()
	if err != nil {
		return "", false, err
	}
	node, found := resolveCurrentSelectedNode(proxies)
	return node, found, nil
}

func resolveCurrentSelectedNode(proxies map[string]model.Proxy) (string, bool) {
	for _, root := range []string{"GLOBAL", "Proxy"} {
		if _, ok := proxies[root]; !ok {
			continue
		}

		if node, found := resolveLeafFromRoot(proxies, root); found {
			return node, true
		}
	}

	return "", false
}

func resolveLeafFromRoot(proxies map[string]model.Proxy, root string) (string, bool) {
	current := root
	visited := make(map[string]struct{}, len(proxies))

	for {
		if _, seen := visited[current]; seen {
			return "", false
		}
		visited[current] = struct{}{}

		proxy, ok := proxies[current]
		if !ok {
			return "", false
		}

		if proxy.Now == "" {
			// 策略组通常带有 all 字段。没有 Now 且仍是策略组时，视为未选择具体节点。
			if len(proxy.All) > 0 {
				return "", false
			}
			return current, true
		}

		current = proxy.Now
	}
}

func boxLine(label, value string, labelWidth, width int) string {
	left := padDisplayRight(label, labelWidth) + ": "
	leftWidth := runewidth.StringWidth(left)
	valueWidth := width - leftWidth
	if valueWidth < 0 {
		valueWidth = 0
	}

	value = fitText(value, valueWidth)
	padding := width - leftWidth - runewidth.StringWidth(value)
	if padding < 0 {
		padding = 0
	}

	return fmt.Sprintf("│ %s%s%s │", left, value, strings.Repeat(" ", padding))
}

func padDisplayRight(s string, width int) string {
	current := runewidth.StringWidth(s)
	if current >= width {
		return s
	}
	return s + strings.Repeat(" ", width-current)
}

func maxDisplayWidth(items []string) int {
	max := 0
	for _, item := range items {
		w := runewidth.StringWidth(item)
		if w > max {
			max = w
		}
	}
	return max
}

func fitText(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}

	rs := []rune(s)
	current := 0
	var out []rune
	for _, r := range rs {
		w := runewidth.RuneWidth(r)
		if current+w > maxWidth-1 {
			break
		}
		out = append(out, r)
		current += w
	}
	return string(out) + "…"
}
