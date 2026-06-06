package dig

import (
	"fmt"
	"testing"
	"time"
)

// TestIsCDNWithMultiDNS 测试使用多DNS服务器检测CDN
func TestIsCDNWithMultiDNS(t *testing.T) {
	// 测试非CDN域名（应该返回单个IP）
	nonCDNDomains := []string{
		"book.fofa.info",
	}

	fmt.Println()
	for _, domain := range nonCDNDomains {
		fmt.Printf("--- 测试域名: %s ---\n", domain)
		dig := NewDig("", 5*time.Second)
		result, err := dig.IsCDN(domain)
		if err != nil {
			fmt.Printf("  错误: %v\n", err)
			continue
		}

		fmt.Printf("  是否CDN: %v\n", result.IsCDN)
		fmt.Printf("  置信度: %.1f%%\n", result.Confidence)
		fmt.Printf("  汇总IP数量: %d\n", len(result.IPs))
		fmt.Printf("  IP列表: %v\n", result.IPs)
		fmt.Printf("  判断原因:\n")
		for _, reason := range result.Reasons {
			fmt.Printf("    - %s\n", reason)
		}
		fmt.Println()
	}
}
