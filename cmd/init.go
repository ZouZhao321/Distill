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

		// 路径使用正斜杠，避免 TOML 解析器将 Windows 反斜杠误读为转义字符
		home := strings.ReplaceAll(storeHome, "\\", "/")
		trash := strings.ReplaceAll(trashPath, "\\", "/")

		configContent := `[core]
    version = "1"
    objects_format = "plain"

[store]
    home = "` + home + `"
    trash_path = "` + trash + `"

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
