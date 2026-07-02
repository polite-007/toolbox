package fofa

import (
	"context"
	"fmt"
	"testing"
)

func TestClient_Host_Basic(t *testing.T) {
	if email == "" || key == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(email, key)

	resp, err := client.Host(&HostRequest{
		Ctx:    context.Background(),
		Host:   "78.48.50.249",
		Detail: false,
	})
	if err != nil {
		t.Fatalf("Host 聚合失败: %v", err)
	}

	fmt.Printf("=== Host 普通模式 ===\n")
	fmt.Printf("Host: %s\n", resp.Host)
	fmt.Printf("IP: %s\n", resp.IP)
	fmt.Printf("ASN: %d\n", resp.ASN)
	fmt.Printf("Org: %s\n", resp.Org)
	fmt.Printf("国家: %s (%s)\n", resp.CountryName, resp.CountryCode)
	fmt.Printf("协议: %v\n", resp.Protocol)
	fmt.Printf("端口: %v\n", resp.Port)
	fmt.Printf("分类: %v\n", resp.Category)
	fmt.Printf("产品: %v\n", resp.Product)
	fmt.Printf("更新时间: %s\n", resp.UpdateTime)
}

func TestClient_Host_Detail(t *testing.T) {
	if email == "" || key == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(email, key)

	resp, err := client.Host(&HostRequest{
		Ctx:    context.Background(),
		Host:   "78.48.50.249",
		Detail: true,
	})
	if err != nil {
		t.Fatalf("Host 聚合（详情）失败: %v", err)
	}

	fmt.Printf("=== Host 详情模式 ===\n")
	fmt.Printf("Host: %s\n", resp.Host)
	fmt.Printf("IP: %s\n", resp.IP)
	fmt.Printf("ASN: %d\n", resp.ASN)
	fmt.Printf("Org: %s\n", resp.Org)
	fmt.Printf("国家: %s (%s)\n", resp.CountryName, resp.CountryCode)
	fmt.Printf("更新时间: %s\n", resp.UpdateTime)
	for _, p := range resp.Ports {
		fmt.Printf("  端口 %d (%s)\n", p.Port, p.Protocol)
		for _, prod := range p.Products {
			fmt.Printf("    - 产品: %s | 分类: %s | 层级: %d | 硬件: %d\n",
				prod.Product, prod.Category, prod.Level, prod.SortHardCode)
		}
	}
}

func TestClient_Host_Empty(t *testing.T) {
	client := NewClient("test", "test")

	if _, err := client.Host(&HostRequest{Host: ""}); err == nil {
		t.Fatal("期望 host 为空时返回错误，但未返回")
	} else {
		fmt.Printf("空 host 错误: %v\n", err)
	}

	if _, err := client.Host(&HostRequest{Host: "   "}); err == nil {
		t.Fatal("期望 host 为空白时返回错误，但未返回")
	}
}
