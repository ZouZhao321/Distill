package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/adapter"
	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/usecase"
	"github.com/ZouZhao321/distill/internal/infra/store"
	"github.com/spf13/cobra"
)

var addName string

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: domain.T(domain.MsgCmdAddShort),
	Long:  domain.T(domain.MsgCmdAddLong),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrCannotAccess), source, err)
		}

		name := addName
		if name == "" {
			name = filepath.Base(source)
		}

		objectStore := store.NewObjectStore(filepath.Join(storeHome, "objects"))
		manifestStore := store.NewManifestStore(
			filepath.Join(storeHome, "manifests"),
			filepath.Join(storeHome, "config", "refs.json"),
		)
		uc := usecase.NewAddAssetUseCase(manifestStore, objectStore)

		switch {
		case info.Mode().IsRegular() && strings.HasSuffix(strings.ToLower(source), ".zip"):
			zipAdapter := adapter.NewZipAdapter(objectStore, true)
			tree, err := zipAdapter.Adapt(source)
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrReadZipFailed), err)
			}
			manifest, err := uc.ExecuteForDirectory(usecase.AddAssetInput{
				Name: name, Tree: tree, Source: source,
			})
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrAddFailed), err)
			}
			fmt.Printf(domain.T(domain.MsgAdded)+"\n", manifest.OriginalName, manifest.FileCount, manifest.TotalSize)

		case info.IsDir():
			dirAdapter := adapter.NewDirAdapter(objectStore, true)
			tree, err := dirAdapter.Adapt(source)
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrReadDirFailed), err)
			}
			manifest, err := uc.ExecuteForDirectory(usecase.AddAssetInput{
				Name: name, Tree: tree, Source: source,
			})
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrAddFailed), err)
			}
			fmt.Printf(domain.T(domain.MsgAdded)+"\n", manifest.OriginalName, manifest.FileCount, manifest.TotalSize)

		default:
			if !info.Mode().IsRegular() {
				return fmt.Errorf(domain.T(domain.MsgErrNotRegularFile), source)
			}
			f, err := os.Open(source)
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrCannotOpen), source, err)
			}
			defer f.Close()

			content, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrReadFailed), source, err)
			}

			manifest, err := uc.Execute(usecase.AddAssetInput{
				Name: name, Content: content, Source: source,
			})
			if err != nil {
				return fmt.Errorf(domain.T(domain.MsgErrAddFailed), err)
			}
			fmt.Printf(domain.T(domain.MsgAdded)+"\n", manifest.OriginalName, manifest.FileCount, manifest.TotalSize)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addName, "as", "n", "", domain.T(domain.MsgFlagAs))
	registerHelpFlag(addCmd)
	rootCmd.AddCommand(addCmd)
}
