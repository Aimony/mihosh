package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [proxy]",
	Short: "检测当前连接信息或测速单个代理节点",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		proxySvc := service.NewProxyService(client, cfg.TestURL, cfg.Timeout)

		// 情况 1: 不带参数，检测当前连接信息
		if len(args) == 0 {
			// 获取链路
			chain, err := proxySvc.GetNodeChain()
			if err != nil {
				fmt.Fprintf(os.Stderr, "获取节点链路失败: %v\n", err)
				os.Exit(1)
			}

			// 获取 IP 信息
			fmt.Printf("正在通过代理 %s 获取出口 IP 信息...\n", cfg.ProxyAddress)
			ipInfo, err := proxySvc.GetIPInfo(cfg.ProxyAddress)
			if err != nil {
				fmt.Fprintf(os.Stderr, "获取 IP 信息失败: %v\n", err)
				fmt.Printf("(请确认配置中的 proxy_address [%s] 是否正确且代理服务已开启)\n", cfg.ProxyAddress)
				os.Exit(1)
			}

			// 格式化输出（固定宽度 + 可视宽度截断，避免中英文混排导致边框错位）
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
			return
		}

		// 情况 2: 带参数，执行原有的测速逻辑
		delay, err := proxySvc.TestProxyDelay(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "测速失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ 节点 '%s' 延迟: %dms\n", args[0], delay)
	},
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
