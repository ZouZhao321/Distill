package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: domain.T(domain.MsgCmdRemoveShort),
	Long:  domain.T(domain.MsgCmdRemoveLong),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)

		// 检查资产是否存在
		ref, err := manifestStore.GetRef(name)
		if err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrAssetNotFound), name)
		}

		// 移除前备份清单到回收站
		manifest, err := manifestStore.GetManifest(ref.Manifest)
		if err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrReadManifestFailed), err)
		}

		if trashPath != "" {
			if err := backupToTrash(manifest, trashPath); err != nil {
				fmt.Fprintf(os.Stderr, domain.T(domain.MsgErrTrashBackupFailed)+"\n", err)
			}
		}

		uc := usecase.NewRemoveUseCase(manifestStore)
		if err := uc.Execute(name); err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrRemoveFailed), err)
		}

		fmt.Printf(domain.T(domain.MsgRemoved)+"\n", name)
		return nil
	},
}

// backupToTrash 将清单 JSON 复制到回收站目录。
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
