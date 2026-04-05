// Package cmd 实现 Distill CLI 的所有命令。
// 包含 init、add、list、checkout、export、remove、gc 七个子命令。
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/spf13/cobra"
)

var (
	storeHome string
	logFormat string
	logLevel  string
	lang      string
	version   = "v0.1.0-dev" // 默认版本，构建时可通过 -ldflags 覆盖
)

var rootCmd = &cobra.Command{
	Use:     "distill",
	Short:   "distill", // 占位，在 applyLang 中动态设置
	Long:    "distill",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// cobra 解析完 flag 后，lang 变量已被赋值
		domain.SetLang(lang)
		applyLangToCommands(cmd)
		setupLogger()
	},
}

// Execute 启动根命令并执行。
func Execute() {
	// 预解析 --lang 参数，在 cobra 帮助渲染之前设置语言
	preParseLang()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// preParseLang 在 cobra 解析前预解析 --lang 参数并应用文案。
func preParseLang() {
	for i, arg := range os.Args[1:] {
		if arg == "--lang" && i+1 < len(os.Args[1:]) {
			lang = os.Args[1:][i+1]
			break
		}
		if len(arg) > 7 && arg[:7] == "--lang=" {
			lang = arg[7:]
			break
		}
	}
	domain.SetLang(lang)

	// 强制 cobra 提前注册内置命令（通常在 Execute 时才注册），
	// 这样 applyLangToCommands 能遍历到它们
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()

	// 强制提前注册 --version flag，使 applyLangToCommands 能更新其 Usage
	rootCmd.InitDefaultVersionFlag()

	applyLangToCommands(rootCmd)
}

// applyLangToCommands 将当前语言的文案应用到根命令和所有子命令。
func applyLangToCommands(root *cobra.Command) {
	root.Short = domain.T(domain.MsgRootShort)
	root.Long = domain.T(domain.MsgRootLong)

	// 覆写帮助模板：替换底部提示语为本地化文本
	helpTip := domain.T(domain.MsgHelpTip)
	root.SetUsageTemplate(strings.ReplaceAll(root.UsageTemplate(),
		`Use "{{.CommandPath}} [command] --help" for more information about a command.`, helpTip))

	// 更新自定义子命令
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

	// 更新 cobra 内置命令的描述
	for _, c := range root.Commands() {
		switch c.Name() {
		case "help":
			c.Short = domain.T(domain.MsgCmdHelpShort)
		case "completion":
			c.Short = domain.T(domain.MsgCmdCompletionShort)
		}
	}

	// 更新全局 PersistentFlags 的 help 文本
	if f := root.PersistentFlags().Lookup("home"); f != nil {
		f.Usage = domain.T(domain.MsgFlagHome)
	}
	if f := root.PersistentFlags().Lookup("log-format"); f != nil {
		f.Usage = domain.T(domain.MsgFlagLogFormat)
	}
	if f := root.PersistentFlags().Lookup("log-level"); f != nil {
		f.Usage = domain.T(domain.MsgFlagLogLevel)
	}
	if f := root.PersistentFlags().Lookup("lang"); f != nil {
		f.Usage = domain.T(domain.MsgFlagLang)
	}
	if f := root.Flags().Lookup("version"); f != nil {
		f.Usage = domain.T(domain.MsgFlagVersion)
	}

	// 自定义版本输出模板，支持本地化
	root.SetVersionTemplate(fmt.Sprintf("%s version {{.Version}}\n", root.Use))

	// 更新各子命令的 LocalFlags help 文本
	updateFlagHelp(addCmd, "as", domain.MsgFlagAs)
	updateFlagHelp(checkoutCmd, "output", domain.MsgFlagOutput)
	updateFlagHelp(checkoutCmd, "overwrite", domain.MsgFlagOverwrite)
	updateFlagHelp(exportCmd, "output", domain.MsgFlagOutput)
	updateFlagHelp(gcCmd, "dry-run", domain.MsgFlagDryRun)
	updateFlagHelp(initCmd, "trash", domain.MsgFlagTrash)
	updateFlagHelp(listCmd, "format", domain.MsgFlagFormat)

	// 更新 cobra 内置 -h/--help flag（根命令 + 所有子命令）
	if f := root.Flags().Lookup("help"); f != nil {
		f.Usage = domain.T(domain.MsgFlagHelp)
	}
	for _, c := range root.Commands() {
		if f := c.Flags().Lookup("help"); f != nil {
			f.Usage = domain.T(domain.MsgFlagHelp)
		}
	}
}

// updateFlagHelp 更新指定命令中某个 flag 的 help 文本。
func updateFlagHelp(cmd *cobra.Command, name string, key domain.MsgKey) {
	if f := cmd.Flags().Lookup(name); f != nil {
		f.Usage = domain.T(key)
	}
}

// registerHelpFlag 为命令提前注册 --help flag，使 applyLangToCommands 能修改其 Usage。
func registerHelpFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("help", "h", false, domain.T(domain.MsgFlagHelp))
}

// rpad 补齐字符串到指定宽度（与 cobra 内部一致）。
func rpad(s string, padding int) string {
	tmpl := fmt.Sprintf(`%%-%ds`, padding)
	return fmt.Sprintf(tmpl, s)
}

// localFlagUsages 返回本地化后的 flag 帮助文本，将 "default" 替换为当前语言文本。
func localFlagUsages(cmd *cobra.Command) string {
	defaultLabel := domain.T(domain.MsgFlagDefault)
	usages := cmd.LocalFlags().FlagUsages()
	return strings.Replace(usages, "(default ", "("+defaultLabel+" ", -1)
}

// inheritedFlagUsages 返回本地化后的全局 flag 帮助文本。
func inheritedFlagUsages(cmd *cobra.Command) string {
	defaultLabel := domain.T(domain.MsgFlagDefault)
	usages := cmd.InheritedFlags().FlagUsages()
	return strings.Replace(usages, "(default ", "("+defaultLabel+" ", -1)
}

// i18nUsageFunc 自定义 UsageFunc，使用本地化的 flag 帮助文本。
func i18nUsageFunc(cmd *cobra.Command) error {
	tmpl := cmd.UsageTemplate()

	// 将模板中的 .LocalFlags.FlagUsages 和 .InheritedFlags.FlagUsages
	// 替换为预处理后的本地化版本
	tmpl = strings.ReplaceAll(tmpl,
		"{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}",
		"{{.LocalFlagUsages | trimTrailingWhitespaces}}")
	tmpl = strings.ReplaceAll(tmpl,
		"{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}",
		"{{.InheritedFlagUsages | trimTrailingWhitespaces}}")

	funcMap := template.FuncMap{
		"rpad":                    rpad,
		"trimTrailingWhitespaces": strings.TrimSpace,
	}

	// 将命令包装为带本地化字段的结构体
	data := struct {
		*cobra.Command
		LocalFlagUsages     string
		InheritedFlagUsages string
	}{
		Command:             cmd,
		LocalFlagUsages:     localFlagUsages(cmd),
		InheritedFlagUsages: inheritedFlagUsages(cmd),
	}

	t, err := template.New("usage").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return err
	}

	fmt.Fprint(cmd.OutOrStdout(), buf.String())
	return nil
}

func init() {
	home, _ := os.UserHomeDir()
	defaultHome := filepath.Join(home, ".distill")
	rootCmd.PersistentFlags().StringVar(&storeHome, "home", defaultHome, domain.T(domain.MsgFlagHome))
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", domain.T(domain.MsgFlagLogFormat))
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", domain.T(domain.MsgFlagLogLevel))
	rootCmd.PersistentFlags().StringVar(&lang, "lang", "zh", domain.T(domain.MsgFlagLang))

	// 替换 cobra 内置 help 命令
	helpCmd := &cobra.Command{
		Use:   "help [command]",
		Short: domain.T(domain.MsgCmdHelpShort),
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.Help()
		},
	}
	rootCmd.SetHelpCommand(helpCmd)

	// 提前注册 help flag 以便 applyLangToCommands 能修改其 Usage
	rootCmd.Flags().BoolP("help", "h", false, domain.T(domain.MsgFlagHelp))

	// 设置自定义 UsageFunc，使用本地化的 "default" 标签
	rootCmd.SetUsageFunc(i18nUsageFunc)
	for _, c := range rootCmd.Commands() {
		c.SetUsageFunc(i18nUsageFunc)
	}
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
