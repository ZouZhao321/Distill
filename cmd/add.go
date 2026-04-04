package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/adapter"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var addName string

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "添加资产到仓库",
	Long:  "将文件、文件夹或 ZIP 包添加到 Distill 仓库。基于 SHA-256 内容寻址，相同内容只存储一份。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("无法访问 %s: %w", source, err)
		}

		name := addName
		if name == "" {
			name = filepath.Base(source)
		}

		objectStore := store.NewObjectStore(filepath.Join(storeHome, "objects"))
		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)
		uc := usecase.NewAddAssetUseCase(manifestStore, objectStore)

		switch {
		case info.Mode().IsRegular() && strings.HasSuffix(strings.ToLower(source), ".zip"):
			zipAdapter := adapter.NewZipAdapter(objectStore, true)
			tree, err := zipAdapter.Adapt(source)
			if err != nil {
				return fmt.Errorf("读取 ZIP 失败: %w", err)
			}
			manifest, err := uc.ExecuteForDirectory(usecase.AddAssetInput{
				Name: name, Tree: tree, Source: source,
			})
			if err != nil {
				return fmt.Errorf("添加失败: %w", err)
			}
			fmt.Printf("已添加: %s (%d 文件, %d bytes)\n", manifest.OriginalName, manifest.FileCount, manifest.TotalSize)

		case info.IsDir():
			dirAdapter := adapter.NewDirAdapter(objectStore, true)
			tree, err := dirAdapter.Adapt(source)
			if err != nil {
				return fmt.Errorf("读取目录失败: %w", err)
			}
			manifest, err := uc.ExecuteForDirectory(usecase.AddAssetInput{
				Name: name, Tree: tree, Source: source,
			})
			if err != nil {
				return fmt.Errorf("添加失败: %w", err)
			}
			fmt.Printf("已添加: %s (%d 文件, %d bytes)\n", manifest.OriginalName, manifest.FileCount, manifest.TotalSize)

		default:
			if !info.Mode().IsRegular() {
				return fmt.Errorf("%s 不是普通文件、目录或 ZIP", source)
			}
			f, err := os.Open(source)
			if err != nil {
				return fmt.Errorf("无法打开 %s: %w", source, err)
			}
			defer f.Close()

			content, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("读取 %s 失败: %w", source, err)
			}

			manifest, err := uc.Execute(usecase.AddAssetInput{
				Name: name, Content: content, Source: source,
			})
			if err != nil {
				return fmt.Errorf("添加失败: %w", err)
			}
			fmt.Printf("已添加: %s (%d 文件, %d bytes)\n", manifest.OriginalName, manifest.FileCount, manifest.TotalSize)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addName, "as", "n", "", "指定资产名称")
	rootCmd.AddCommand(addCmd)
}
