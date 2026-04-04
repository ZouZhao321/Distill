package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
)

var gcDryRun bool

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "垃圾回收",
	Long:  "清理未被任何清单引用的孤立对象，释放磁盘空间。",
	RunE: func(cmd *cobra.Command, args []string) error {
		objectStore := store.NewObjectStore(filepath.Join(storeHome, "objects"))
		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)

		uc := usecase.NewGCUseCase(manifestStore, objectStore)

		if gcDryRun {
			orphans, err := uc.ExecuteDryRun()
			if err != nil {
				return fmt.Errorf("GC 预检失败: %w", err)
			}
			if len(orphans) == 0 {
				fmt.Println("没有发现孤立对象。")
			} else {
				fmt.Printf("发现 %d 个孤立对象:\n", len(orphans))
				for _, h := range orphans {
					fmt.Printf("  %s\n", h)
				}
			}
			return nil
		}

		cleaned, err := uc.Execute()
		if err != nil {
			return fmt.Errorf("GC 执行失败: %w", err)
		}

		if cleaned == 0 {
			fmt.Println("仓库已是干净状态，无需清理。")
		} else {
			fmt.Printf("已清理 %d 个孤立对象。\n", cleaned)
		}
		return nil
	},
}

func init() {
	gcCmd.Flags().BoolVar(&gcDryRun, "dry-run", false, "仅列出孤立对象，不删除")
	rootCmd.AddCommand(gcCmd)
}
