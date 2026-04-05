// Package cmd 实现 Distill CLI 的所有命令。
// 包含 init、add、list、checkout、export、remove、gc 七个子命令。
package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

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

// Execute 启动根命令并执行。
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

// setupLogger 根据 CLI 参数初始化全局 slog 日志。
// 日志同时输出到 stderr 和仓库 log 目录下的日志文件（如果目录存在）。
func setupLogger() {
	level := parseLogLevel(logLevel)

	// 日志文件输出目标：仅在 log 目录存在时写入文件
	logDir := filepath.Join(storeHome, "log")
	var writers []io.Writer
	writers = append(writers, os.Stderr)

	logFile, err := openLogFile(logDir)
	if err == nil && logFile != nil {
		writers = append(writers, logFile)
	}

	var handler slog.Handler
	writer := io.MultiWriter(writers...)
	switch logFormat {
	case "json":
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}

// openLogFile 在 logDir 下创建或打开当天的日志文件。
// 如果 logDir 不存在，返回 nil（静默跳过文件日志）。
func openLogFile(logDir string) (*os.File, error) {
	info, err := os.Stat(logDir)
	if err != nil || !info.IsDir() {
		return nil, nil // log 目录不存在，跳过文件日志
	}

	filename := filepath.Join(logDir, fmt.Sprintf("distill-%s.log", time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// parseLogLevel 将字符串日志级别转换为 slog.Level。
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
