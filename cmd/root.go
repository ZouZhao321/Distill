package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	storeHome string
	logFormat string
	logLevel  string
)

var rootCmd = &cobra.Command{
	Use:   "distill",
	Short: "Distill - 资产管理 CLI 工具",
	Long:  "Distill 是一个 Go 语言 CLI 工具，用于资产的导入、管理和导出。基于内容寻址存储实现物理级去重。",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogger()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	home, _ := os.UserHomeDir()
	defaultHome := filepath.Join(home, ".distill")
	rootCmd.PersistentFlags().StringVar(&storeHome, "home", defaultHome, "仓库路径")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "日志格式 (text|json)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "日志级别 (debug|info|warn|error)")
}

// setupLogger initializes the global slog logger based on CLI flags.
func setupLogger() {
	level := parseLogLevel(logLevel)

	var handler slog.Handler
	switch logFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}

// parseLogLevel converts a string log level to slog.Level.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
