package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/spf13/cobra"
)

var trashPath string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: domain.T(domain.MsgCmdInitShort),
	Long:  domain.T(domain.MsgCmdInitLong),
	RunE: func(cmd *cobra.Command, args []string) error {
		storeHome := resolveStoreHome()

		dirs := []string{
			filepath.Join(storeHome, "objects"),
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config"),
			filepath.Join(storeHome, "log"),
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrCreateDirFailed), d, err)
			}
		}

		// 使用 Config struct 生成配置，确保格式一致
		config := domain.DefaultConfig()
		config.Store.Home = strings.ReplaceAll(storeHome, "\\", "/")
		config.Store.TrashPath = strings.ReplaceAll(trashPath, "\\", "/")
		config.Lang = resolveLang()

		configPath := filepath.Join(storeHome, "config", "config.toml")
		if err := domain.SaveConfig(config, configPath); err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrWriteConfigFailed), err)
		}

		refsPath := filepath.Join(storeHome, "config", "refs.json")
		if err := os.WriteFile(refsPath, []byte("{}\n"), 0644); err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrWriteRefsFailed), err)
		}

		fmt.Printf(domain.T(domain.MsgInited)+"\n", storeHome)
		fmt.Printf(domain.T(domain.MsgTrashPath)+"\n", trashPath)
		return nil
	},
}

func init() {
	defaultTrash, _ := os.UserHomeDir()
	initCmd.Flags().StringVar(&trashPath, "trash", filepath.Join(defaultTrash, ".distill-trash"), domain.T(domain.MsgFlagTrash))
	registerHelpFlag(initCmd)
	rootCmd.AddCommand(initCmd)
}
