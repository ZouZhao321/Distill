package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/spf13/cobra"
)

var trashPath string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: domain.T(domain.MsgCmdInitShort),
	Long:  domain.T(domain.MsgCmdInitLong),
	RunE: func(cmd *cobra.Command, args []string) error {
		storeHome := resolveStoreHome()
		lang := resolveLang()

		uc := usecase.NewInitUseCase()
		_, err := uc.Execute(usecase.InitInput{
			StoreHome: storeHome,
			TrashPath: trashPath,
			Lang:      lang,
		})
		if err != nil {
			if err == usecase.ErrAlreadyInitialized {
				return fmt.Errorf("%s: %w", domain.T(domain.MsgErrCreateDirFailed, domain.P{"Path": storeHome}), err)
			}
			return err
		}

		fmt.Println(domain.T(domain.MsgInited, domain.P{"Path": storeHome}))
		fmt.Println(domain.T(domain.MsgTrashPath, domain.P{"Path": trashPath}))
		return nil
	},
}

func init() {
	defaultTrash, _ := os.UserHomeDir()
	initCmd.Flags().StringVar(&trashPath, "trash", filepath.Join(defaultTrash, ".distill-trash"), domain.T(domain.MsgFlagTrash))
	registerHelpFlag(initCmd)
	rootCmd.AddCommand(initCmd)
}
