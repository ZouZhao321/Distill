package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var trashPath string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 Distill 仓库",
	Long:  "在指定路径创建 Distill 仓库目录结构、默认配置文件和空的引用索引。",
	RunE: func(cmd *cobra.Command, args []string) error {
		dirs := []string{
			filepath.Join(storeHome, "objects"),
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config"),
			filepath.Join(storeHome, "log"),
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf("创建目录 %s 失败: %w", d, err)
			}
		}

		configContent := `[core]
    version = "1"
    objects_format = "plain"

[store]
    home = "` + storeHome + `"
    trash_path = "` + trashPath + `"

[checkout]
    overwrite = "ask"

[log]
    format = "text"
    level = "info"

[normalize]
    crlf_to_lf = true
`
		configPath := filepath.Join(storeHome, "config", "config.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return fmt.Errorf("写入配置文件失败: %w", err)
		}

		refsPath := filepath.Join(storeHome, "config", "refs.json")
		if err := os.WriteFile(refsPath, []byte("{}\n"), 0644); err != nil {
			return fmt.Errorf("写入引用文件失败: %w", err)
		}

		fmt.Printf("Distill 仓库初始化完成: %s\n", storeHome)
		fmt.Printf("回收站路径: %s\n", trashPath)
		return nil
	},
}

func init() {
	defaultTrash, _ := os.UserHomeDir()
	initCmd.Flags().StringVar(&trashPath, "trash", filepath.Join(defaultTrash, ".distill-trash"), "回收站路径")
	rootCmd.AddCommand(initCmd)
}
