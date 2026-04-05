package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var listFormat string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: domain.T(domain.MsgCmdListShort),
	Long:  domain.T(domain.MsgCmdListLong),
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)

		uc := usecase.NewListAssetsUseCase(manifestStore)
		items, err := uc.Execute()
		if err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrListFailed), err)
		}

		if len(items) == 0 {
			fmt.Println(domain.T(domain.MsgListEmpty))
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
	listCmd.Flags().StringVar(&listFormat, "format", "table", domain.T(domain.MsgFlagFormat))
	rootCmd.AddCommand(listCmd)
}

func printJSON(items []usecase.ListItem) error {
	// 不依赖外部库的简单 JSON 输出
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
