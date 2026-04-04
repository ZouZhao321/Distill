package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
)

var listFormat string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有资产",
	Long:  "显示仓库中所有已导入的资产列表。",
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)

		uc := usecase.NewListAssetsUseCase(manifestStore)
		items, err := uc.Execute()
		if err != nil {
			return fmt.Errorf("列表查询失败: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("仓库为空，使用 distill add 添加资产")
			return nil
		}

		if listFormat == "json" {
			return printJSON(items)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tFILES\tSIZE\tCREATED")
		for _, item := range items {
			fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", item.Name, item.FileCount, item.TotalSize, item.CreatedAt)
		}
		w.Flush()
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", "table", "输出格式 (table|json)")
	rootCmd.AddCommand(listCmd)
}

func printJSON(items []usecase.ListItem) error {
	// Simple JSON output without external dependency
	fmt.Println("[")
	for i, item := range items {
		sep := ","
		if i == len(items)-1 {
			sep = ""
		}
		fmt.Printf("  {\"name\": %q, \"hash\": %q, \"file_count\": %d, \"total_size\": %d, \"created_at\": %q}%s\n",
			item.Name, item.Hash, item.FileCount, item.TotalSize, item.CreatedAt, sep)
	}
	fmt.Println("]")
	return nil
}
