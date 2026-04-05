package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: domain.T(domain.MsgCmdExportShort),
	Long:  domain.T(domain.MsgCmdExportLong),
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
			return fmt.Errorf(domain.T(domain.MsgErrExportFailed), err)
		}

		fmt.Printf(domain.T(domain.MsgExported)+"\n", name, outputPath)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", domain.T(domain.MsgFlagOutput))
	registerHelpFlag(exportCmd)
	rootCmd.AddCommand(exportCmd)
}
