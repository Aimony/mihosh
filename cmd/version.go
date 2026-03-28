package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// 通过 -ldflags "-X cmd.Version=x.y.z -X cmd.Commit=abc -X cmd.Date=2006-01-02" 注入
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionOutput string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示 mihosh 的版本号、提交哈希和构建日期。`,
	Example: `  mihosh version
  mihosh version --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := parseOutputFormat(versionOutput)
		if err != nil {
			return wrapParameterError(err)
		}
		return renderVersion(os.Stdout, format)
	},
}

func init() {
	versionCmd.Flags().StringVar(&versionOutput, "output", string(outputFormatPlain), "输出格式: json|plain")
}

func renderVersion(w io.Writer, format outputFormat) error {
	switch format {
	case outputFormatJSON:
		return writeJSON(w, map[string]string{
			"version": Version,
			"commit":  Commit,
			"date":    Date,
		})
	case outputFormatTable:
		tw := newTabWriter(w)
		fmt.Fprintln(tw, "KEY\tVALUE")
		fmt.Fprintf(tw, "VERSION\t%s\n", Version)
		fmt.Fprintf(tw, "COMMIT\t%s\n", Commit)
		fmt.Fprintf(tw, "DATE\t%s\n", Date)
		return tw.Flush()
	default:
		fmt.Fprintf(w, "mihosh %s (commit: %s, built: %s)\n", Version, Commit, Date)
		return nil
	}
}
