package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var exportOutput string
var exportOverwrite string

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

		overwrite := exportOverwrite
		if overwrite == "" {
			overwrite = "ask"
		}

		home := resolveStoreHome()
		manifestStore := store.NewManifestStore(
			filepath.Join(home, "manifests"),
			filepath.Join(home, "config", "refs.json"),
		)
		objectStore := store.NewObjectStore(filepath.Join(home, "objects"))

		uc := usecase.NewExportUseCase(manifestStore, objectStore)
		err := uc.Execute(name, outputPath, overwrite)
		if err != nil {
			if err == domain.ErrAlreadyExists && overwrite == "ask" {
				fmt.Printf(domain.T(domain.MsgExportFileExists)+"\n", outputPath)
				fmt.Print(domain.T(domain.MsgExportOverwritePrompt))
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))

				if answer == "y" || answer == "yes" {
					err = uc.Execute(name, outputPath, "force")
					if err != nil {
						return fmt.Errorf(domain.T(domain.MsgErrExportFailed), err)
					}
				} else {
					fmt.Println(domain.T(domain.MsgExportSkipped))
					return nil
				}
			} else {
				return fmt.Errorf(domain.T(domain.MsgErrExportFailed), err)
			}
		}

		fmt.Printf(domain.T(domain.MsgExported)+"\n", name, outputPath)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", domain.T(domain.MsgFlagOutput))
	exportCmd.Flags().StringVar(&exportOverwrite, "overwrite", "ask", domain.T(domain.MsgFlagOverwrite))
	registerHelpFlag(exportCmd)
	rootCmd.AddCommand(exportCmd)
}
