package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: domain.T(domain.MsgCmdConfigShort),
	Long:  domain.T(domain.MsgCmdConfigLong),
}

// --- show 子命令 ---

var configShowCmd = &cobra.Command{
	Use:   "show [key]",
	Short: domain.T(domain.MsgCmdConfigShowShort),
	Long:  domain.T(domain.MsgCmdConfigShowLong),
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadAppConfig()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			// 显示所有配置（含说明和可选值）
			fmt.Println(domain.FormatAllConfig(config))
			return nil
		}

		// 显示单个配置项（含说明和可选值）
		key := args[0]
		value, err := domain.GetConfigValue(config, key)
		if err != nil {
			return err
		}
		fmt.Println(domain.FormatConfigItem(key, value))
		return nil
	},
}

// --- get 子命令（show 的别名） ---

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: domain.T(domain.MsgCmdConfigGetShort),
	Long:  domain.T(domain.MsgCmdConfigGetLong),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadAppConfig()
		if err != nil {
			return err
		}

		value, err := domain.GetConfigValue(config, args[0])
		if err != nil {
			return err
		}
		fmt.Println(domain.FormatConfigItem(args[0], value))
		return nil
	},
}

// --- set 子命令 ---

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: domain.T(domain.MsgCmdConfigSetShort),
	Long:  domain.T(domain.MsgCmdConfigSetLong),
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		config, err := loadAppConfig()
		if err != nil {
			return err
		}

		if err := domain.SetConfigValue(config, key, value); err != nil {
			return err
		}

		configPath := filepath.Join(resolveStoreHome(), "config", "config.toml")
		if err := domain.SaveConfig(config, configPath); err != nil {
			return fmt.Errorf(domain.T(domain.MsgErrWriteConfigFailed), err)
		}

		fmt.Printf("%s=%s\n", key, value)
		return nil
	},
}

func init() {
	registerHelpFlag(configShowCmd)
	registerHelpFlag(configGetCmd)
	registerHelpFlag(configSetCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

// loadAppConfig 加载应用配置。
// 先尝试从配置文件加载，如果文件不存在则返回默认配置。
func loadAppConfig() (*domain.Config, error) {
	home := resolveStoreHome()
	config, err := domain.LoadConfigByHome(home)
	if err != nil {
		return nil, err
	}

	// 如果配置文件不存在（空 Config），返回默认值
	if config.Store.Home == "" {
		return domain.DefaultConfig(), nil
	}
	return config, nil
}
