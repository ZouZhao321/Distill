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

var checkoutOutput string
var checkoutOverwrite string

var checkoutCmd = &cobra.Command{
	Use:   "checkout <name>",
	Short: domain.T(domain.MsgCmdCheckoutShort),
	Long:  domain.T(domain.MsgCmdCheckoutLong),
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

		home := resolveStoreHome()
		manifestStore := store.NewManifestStore(
			filepath.Join(home, "manifests"),
			filepath.Join(home, "config", "refs.json"),
		)
		objectStore := store.NewObjectStore(filepath.Join(home, "objects"))

		uc := usecase.NewCheckoutUseCase(manifestStore, objectStore)
		err := uc.Execute(name, outputDir, overwrite)
		if err != nil {
			if err == domain.ErrAlreadyExists && overwrite == "ask" {
				fmt.Printf(domain.T(domain.MsgCheckoutFileExists)+"\n", outputDir)
				fmt.Print(domain.T(domain.MsgCheckoutOverwritePrompt))
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))

				if answer == "y" || answer == "yes" {
					err = uc.Execute(name, outputDir, "force")
					if err != nil {
						return fmt.Errorf(domain.T(domain.MsgErrCheckoutFailed), err)
					}
				} else {
					fmt.Println(domain.T(domain.MsgCheckoutSkipped))
					return nil
				}
			} else {
				return fmt.Errorf(domain.T(domain.MsgErrCheckoutFailed), err)
			}
		}

		fmt.Printf(domain.T(domain.MsgCheckedOut)+"\n", name, outputDir)
		return nil
	},
}

func init() {
	checkoutCmd.Flags().StringVarP(&checkoutOutput, "output", "o", "", domain.T(domain.MsgFlagOutput))
	checkoutCmd.Flags().StringVar(&checkoutOverwrite, "overwrite", "skip", domain.T(domain.MsgFlagOverwrite))
	registerHelpFlag(checkoutCmd)
	rootCmd.AddCommand(checkoutCmd)
}
