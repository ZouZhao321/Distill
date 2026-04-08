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
				return fmt.Errorf(domain.T(domain.MsgErrCreateDirFailed), storeHome, err)
			}
			return err
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
