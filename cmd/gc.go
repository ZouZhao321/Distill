package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var gcDryRun bool

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: domain.T(domain.MsgCmdGcShort),
	Long:  domain.T(domain.MsgCmdGcLong),
	RunE: func(cmd *cobra.Command, args []string) error {
		objectStore := store.NewObjectStore(filepath.Join(storeHome, "objects"))
		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)

		uc := usecase.NewGCUseCase(manifestStore, objectStore)

		if gcDryRun {
			orphans, err := uc.ExecuteDryRun()
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrGcDryRunFailed), err)
			}
			if len(orphans) == 0 {
				fmt.Println(domain.T(domain.MsgGcNoOrphans))
			} else {
				fmt.Printf(domain.T(domain.MsgGcOrphanList)+"\n", len(orphans))
				for _, h := range orphans {
					fmt.Printf("  %s\n", h)
				}
			}
			return nil
		}

		cleaned, err := uc.Execute()
		if err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrGcFailed), err)
		}

		if cleaned == 0 {
			fmt.Println(domain.T(domain.MsgGcAlreadyClean))
		} else {
			fmt.Printf(domain.T(domain.MsgGcClean)+"\n", cleaned)
		}
		return nil
	},
}

func init() {
	gcCmd.Flags().BoolVar(&gcDryRun, "dry-run", false, domain.T(domain.MsgFlagDryRun))
	registerHelpFlag(gcCmd)
	rootCmd.AddCommand(gcCmd)
}
