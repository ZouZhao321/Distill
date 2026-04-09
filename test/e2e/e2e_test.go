//go:build e2e

// Package e2e 提供 Distill CLI 端到端黑盒测试。
// 测试通过 exec.Command 调用编译后的 distill 二进制，验证完整集成行为。
//
// 运行方式:
//
//	go test ./test/e2e/ -tags=e2e -v
package e2e
