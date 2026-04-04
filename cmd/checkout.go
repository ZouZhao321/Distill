package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
)

var checkoutOutput string
var checkoutOverwrite string

var checkoutCmd = &cobra.Command{
	Use:   "checkout <name>",
	Short: "从仓库还原资产到目录",
	Long:  "将指定资产从仓库还原到目标目录，保留原始目录结构。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		outputDir := checkoutOutput
		if outputDir == "" {
			outputDir = name
		}

		overwrite := checkoutOverwrite
		if overwrite == "" {
			overwrite = "skip"
		}

		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)
		objectStore := store.NewObjectStore(filepath.Join(storeHome, "objects"))

		uc := usecase.NewCheckoutUseCase(manifestStore, objectStore)
		err := uc.Execute(name, outputDir, overwrite)
		if err != nil {
			if err == domain.ErrAlreadyExists && overwrite == "ask" {
				// Ask user for confirmation
				fmt.Printf("文件已存在: %s\n", outputDir)
				fmt.Print("是否覆盖？(y/N): ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))

				if answer == "y" || answer == "yes" {
					err = uc.Execute(name, outputDir, "force")
					if err != nil {
						return fmt.Errorf("还原失败: %w", err)
					}
				} else {
					fmt.Println("已跳过。")
					return nil
				}
			} else {
				return fmt.Errorf("还原失败: %w", err)
			}
		}

		fmt.Printf("已还原: %s -> %s\n", name, outputDir)
		return nil
	},
}

func init() {
	checkoutCmd.Flags().StringVarP(&checkoutOutput, "output", "o", "", "输出目录路径")
	checkoutCmd.Flags().StringVar(&checkoutOverwrite, "overwrite", "skip", "覆盖策略 (skip|force|ask)")
	rootCmd.AddCommand(checkoutCmd)
}
