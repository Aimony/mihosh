package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		configSvc := service.NewConfigService()
		if err := configSvc.InitConfig(); err != nil {
			return wrapConfigError(fmt.Errorf("初始化配置失败: %w", err))
		}
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show [--output json|table|plain]",
	Short: "显示当前配置（支持多种输出格式）",
	Long: `显示当前配置内容。

可通过 --output 选择输出格式：
  plain  人类可读文本（默认）
  table  表格输出
  json   结构化 JSON 输出`,
	Example: `  mihosh config show
  mihosh config show --output table
  mihosh config show --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := parseOutputFormat(configShowOutput)
		if err != nil {
			return wrapParameterError(err)
		}

		configSvc := service.NewConfigService()
		cfg, err := configSvc.LoadConfig()
		if err != nil {
			return wrapConfigError(fmt.Errorf("加载配置失败: %w", err))
		}

		configDir, _ := config.GetConfigDir()
		configPath := filepath.Join(configDir, "config.yaml")

		if err := renderConfigShow(os.Stdout, cfg, configPath, format); err != nil {
			return fmt.Errorf("渲染输出失败: %w", err)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项",
	Long: `设置配置项

可用的配置项:
  api-address  - mihosh API 地址 (例如: http://127.0.0.1:9090)
  secret       - API 密钥
  test-url     - 测速 URL (例如: http://www.gstatic.com/generate_204)
  timeout      - 超时时间，单位毫秒 (例如: 5000)
  proxy-address - HTTP 代理地址 (例如: http://127.0.0.1:7890)

示例:
  mihosh config set api-address http://127.0.0.1:9090
  mihosh config set secret your-secret-here
  mihosh config set test-url http://www.google.com/generate_204
  mihosh config set timeout 3000
  mihosh config set proxy-address http://127.0.0.1:7890`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		configSvc := service.NewConfigService()
		if err := configSvc.SetConfigValue(key, value); err != nil {
			if isConfigSetValidationError(err) {
				return wrapParameterError(fmt.Errorf("设置配置失败: %w", err))
			}
			return wrapConfigError(fmt.Errorf("设置配置失败: %w", err))
		}

		fmt.Printf("✓ 已设置 %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configShowCmd.Flags().StringVar(&configShowOutput, "output", string(outputFormatPlain), "输出格式: json|table|plain")
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

var configShowOutput string

func isConfigSetValidationError(err error) bool {
	msg := strings.TrimSpace(err.Error())
	return strings.Contains(msg, "未知的配置项:") || strings.Contains(msg, "timeout 必须是数字:")
}

func renderConfigShow(w io.Writer, cfg *config.Config, configPath string, format outputFormat) error {
	switch format {
	case outputFormatJSON:
		payload := struct {
			APIAddress   string `json:"api_address"`
			Secret       string `json:"secret"`
			TestURL      string `json:"test_url"`
			TimeoutMS    int    `json:"timeout_ms"`
			ProxyAddress string `json:"proxy_address"`
			ConfigFile   string `json:"config_file"`
		}{
			APIAddress:   cfg.APIAddress,
			Secret:       utils.MaskSecret(cfg.Secret),
			TestURL:      cfg.TestURL,
			TimeoutMS:    cfg.Timeout,
			ProxyAddress: cfg.ProxyAddress,
			ConfigFile:   configPath,
		}
		return writeJSON(w, payload)
	case outputFormatTable:
		tw := newTabWriter(w)
		fmt.Fprintln(tw, "KEY\tVALUE")
		fmt.Fprintf(tw, "API_ADDRESS\t%s\n", cfg.APIAddress)
		fmt.Fprintf(tw, "SECRET\t%s\n", utils.MaskSecret(cfg.Secret))
		fmt.Fprintf(tw, "TEST_URL\t%s\n", cfg.TestURL)
		fmt.Fprintf(tw, "TIMEOUT_MS\t%d\n", cfg.Timeout)
		fmt.Fprintf(tw, "PROXY_ADDRESS\t%s\n", cfg.ProxyAddress)
		fmt.Fprintf(tw, "CONFIG_FILE\t%s\n", configPath)
		return tw.Flush()
	case outputFormatPlain:
		fmt.Fprintln(w, "当前配置:")
		fmt.Fprintf(w, "  API 地址: %s\n", cfg.APIAddress)
		fmt.Fprintf(w, "  密钥:     %s\n", utils.MaskSecret(cfg.Secret))
		fmt.Fprintf(w, "  测速 URL: %s\n", cfg.TestURL)
		fmt.Fprintf(w, "  超时:     %dms\n", cfg.Timeout)
		fmt.Fprintf(w, "  代理地址: %s\n", cfg.ProxyAddress)
		fmt.Fprintf(w, "\n配置文件位置: %s\n", configPath)
		return nil
	default:
		return fmt.Errorf("不支持的输出格式: %s", format)
	}
}
