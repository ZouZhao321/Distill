package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "导出资产为 ZIP 压缩包",
	Long:  "将指定资产从仓库打包导出为 ZIP 文件。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		outputPath := exportOutput
		if outputPath == "" {
			outputPath = name + ".zip"
		}

		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)
		objectStore := store.NewObjectStore(filepath.Join(storeHome, "objects"))

		uc := usecase.NewExportUseCase(manifestStore, objectStore)
		err := uc.Execute(name, outputPath)
		if err != nil {
			return fmt.Errorf("导出失败: %w", err)
		}

		fmt.Printf("已导出: %s -> %s\n", name, outputPath)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "输出 ZIP 文件路径")
	rootCmd.AddCommand(exportCmd)
}
