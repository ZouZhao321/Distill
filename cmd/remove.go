package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "移除资产",
	Long:  "从仓库中移除指定名称的资产。清单和引用将被删除，对象数据保留，等待 GC 清理。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)

		// Check if asset exists first for better error message
		ref, err := manifestStore.GetRef(name)
		if err != nil {
			return fmt.Errorf("资产 %q 不存在", name)
		}

		// Backup manifest to trash before removal
		manifest, err := manifestStore.GetManifest(ref.Manifest)
		if err != nil {
			return fmt.Errorf("无法读取清单: %w", err)
		}

		if trashPath != "" {
			if err := backupToTrash(manifest, trashPath); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 回收站备份失败: %v\n", err)
			}
		}

		uc := usecase.NewRemoveUseCase(manifestStore)
		if err := uc.Execute(name); err != nil {
			return fmt.Errorf("移除失败: %w", err)
		}

		fmt.Printf("已移除: %s\n", name)
		return nil
	},
}

// backupToTrash copies the manifest JSON to the trash directory.
func backupToTrash(manifest *domain.Manifest, trashPath string) error {
	if err := os.MkdirAll(trashPath, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	trashFile := filepath.Join(trashPath, manifest.Hash+"manifest.json")
	return os.WriteFile(trashFile, data, 0644)
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
