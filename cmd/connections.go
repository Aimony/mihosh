package cmd

import (
	"fmt"
	"os"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/spf13/cobra"
)

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
		connSvc := service.NewConnectionService(client)

		conns, err := connSvc.GetConnections()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取连接失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("活跃连接数: %d\n", len(conns.Connections))
		fmt.Printf("上传总量: %s\n", utils.FormatBytes(conns.UploadTotal))
		fmt.Printf("下载总量: %s\n", utils.FormatBytes(conns.DownloadTotal))
		fmt.Println("\n连接列表:")

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
