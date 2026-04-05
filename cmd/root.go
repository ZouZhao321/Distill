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

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/spf13/cobra"
)

var (
	storeHome string
	logFormat string
	logLevel  string
	lang      string
)

var rootCmd = &cobra.Command{
	Use:   "distill",
	Short: domain.T(domain.MsgRootShort),
	Long:  domain.T(domain.MsgRootLong),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		domain.SetLang(lang)
		// 设置语言后重新绑定根命令的 Short/Long
		cmd.Short = domain.T(domain.MsgRootShort)
		cmd.Long = domain.T(domain.MsgRootLong)
		setupLogger()
	},
}

// Execute 启动根命令并执行。
func Execute() {
	// 预解析 --lang 参数，在 cobra 帮助渲染之前设置语言
	for i, arg := range os.Args[1:] {
		if arg == "--lang" && i+1 < len(os.Args[1:]) {
			domain.SetLang(os.Args[1:][i+1])
			break
		}
		if len(arg) > 7 && arg[:7] == "--lang=" {
			domain.SetLang(arg[7:])
			break
		}
	}

	// 设置语言后更新所有命令的文案
	applyLangToCommands()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// applyLangToCommands 将当前语言的文案应用到根命令和所有子命令。
func applyLangToCommands() {
	rootCmd.Short = domain.T(domain.MsgRootShort)
	rootCmd.Long = domain.T(domain.MsgRootLong)

	shortMap := map[*cobra.Command]domain.MsgKey{
		addCmd:      domain.MsgCmdAddShort,
		checkoutCmd: domain.MsgCmdCheckoutShort,
		exportCmd:   domain.MsgCmdExportShort,
		gcCmd:       domain.MsgCmdGcShort,
		initCmd:     domain.MsgCmdInitShort,
		listCmd:     domain.MsgCmdListShort,
		removeCmd:   domain.MsgCmdRemoveShort,
	}
	longMap := map[*cobra.Command]domain.MsgKey{
		addCmd:      domain.MsgCmdAddLong,
		checkoutCmd: domain.MsgCmdCheckoutLong,
		exportCmd:   domain.MsgCmdExportLong,
		gcCmd:       domain.MsgCmdGcLong,
		initCmd:     domain.MsgCmdInitLong,
		listCmd:     domain.MsgCmdListLong,
		removeCmd:   domain.MsgCmdRemoveLong,
	}
	for cmd, key := range shortMap {
		cmd.Short = domain.T(key)
	}
	for cmd, key := range longMap {
		cmd.Long = domain.T(key)
	}
}

func init() {
	home, _ := os.UserHomeDir()
	defaultHome := filepath.Join(home, ".distill")
	rootCmd.PersistentFlags().StringVar(&storeHome, "home", defaultHome, domain.T(domain.MsgFlagHome))
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", domain.T(domain.MsgFlagLogFormat))
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", domain.T(domain.MsgFlagLogLevel))
	rootCmd.PersistentFlags().StringVar(&lang, "lang", "zh", domain.T(domain.MsgFlagLang))
}

// setupLogger 根据配置文件和 CLI 参数初始化全局 slog 日志。
// 优先级：CLI flag > config.toml > 默认值。
// 仓库已初始化时日志只写入文件；未初始化时 fallback 到 stderr。
func setupLogger() {
	// 尝试从 config.toml 加载日志配置
	configPath := filepath.Join(storeHome, "config", "config.toml")
	config, err := domain.LoadConfig(configPath)
	if err != nil {
		config = &domain.Config{}
	}

	// CLI flag 为空时使用配置文件值，配置文件值也为空时使用默认值
	format := logFormat
	if format == "" {
		format = config.Log.Format
	}
	if format == "" {
		format = "text"
	}

	levelStr := logLevel
	if levelStr == "" {
		levelStr = config.Log.Level
	}
	if levelStr == "" {
		levelStr = "info"
	}

	level := parseLogLevel(levelStr)

	// 确定日志输出目标：优先写文件，仅在文件不可用时 fallback 到 stderr
	logDir := filepath.Join(storeHome, "log")
	var writer io.Writer
	logFile, err := openLogFile(logDir)
	if err == nil && logFile != nil {
		writer = logFile
	} else {
		writer = os.Stderr
		if err != nil {
			fmt.Fprintf(os.Stderr, domain.T(domain.MsgWarnLogOpenFailed), err)
		}
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}

// openLogFile 在 logDir 下创建或打开当天的日志文件。
// 如果 logDir 不存在，返回 nil（调用方将 fallback 到 stderr）。
func openLogFile(logDir string) (*os.File, error) {
	info, err := os.Stat(logDir)
	if err != nil || !info.IsDir() {
		return nil, nil // log 目录不存在，fallback 到 stderr
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
