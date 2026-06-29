package fofa

import (
	"context"
	"fmt"
	"os"
	"testing"
)

var (
	email = os.Getenv("FOFA_EMAIL")
	key   = os.Getenv("FOFA_KEY")
)

func TestClient_Account(t *testing.T) {
	if email == "" || key == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(email, key)

	resp, err := client.Account(context.Background())
	if err != nil {
		t.Fatalf("获取账号信息失败: %v", err)
	}

	fmt.Printf("=== 账号信息 ===\n")
	fmt.Printf("邮箱: %s\n", resp.Email)
	fmt.Printf("用户名: %s\n", resp.Username)
	fmt.Printf("类别: %s\n", resp.Category)
	fmt.Printf("F币: %d\n", resp.Fcoin)
	fmt.Printf("F点数: %d\n", resp.FofaPoint)
	fmt.Printf("剩余免费点数: %d\n", resp.RemainFreePoint)
	fmt.Printf("剩余API查询次数: %d\n", resp.RemainAPIQuery)
	fmt.Printf("剩余API数据条数: %d\n", resp.RemainAPIData)
	fmt.Printf("是否VIP: %v\n", resp.IsVip)
	fmt.Printf("VIP等级: %d\n", resp.VipLevel)
	fmt.Printf("是否已认证: %v\n", resp.IsVerified)
	fmt.Printf("头像: %s\n", resp.Avatar)
	fmt.Printf("消息: %s\n", resp.Message)
	fmt.Printf("fofacli版本: %s\n", resp.FofacliVer)
	fmt.Printf("FOFA服务端: %v\n", resp.FofaServer)
}
