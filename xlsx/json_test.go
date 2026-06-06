package xlsx

import (
	"encoding/json"
	"testing"
)

// 示例结构体1：用户信息
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Password string `json:"password"` // 敏感字段，通常不导出
}

// 示例结构体2：产品信息
type Product struct {
	ID          int     `json:"id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

// 示例结构体3：带omitempty标签
type Order struct {
	OrderID    string  `json:"order_id"`
	CustomerID int     `json:"customer_id"`
	Amount     float64 `json:"amount,omitempty"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
}

func TestExportModels_BasicUsage(t *testing.T) {
	// 准备测试数据
	users := []*User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25, Password: "secret123"},
		{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 30, Password: "secret456"},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 28, Password: "secret789"},
	}

	// 只导出 id, name, email 字段
	fields := []string{"id", "name", "email"}

	// 调用 ExportModels
	result, err := ExportModels(users, fields)
	if err != nil {
		t.Fatalf("ExportModels failed: %v", err)
	}

	// 验证结果
	var output []map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// 验证结果数量
	if len(output) != 3 {
		t.Errorf("Expected 3 records, got %d", len(output))
	}

	// 验证第一个记录的字段
	first := output[0]
	if first["id"] != float64(1) { // JSON 数字会被解析为 float64
		t.Errorf("Expected id=1, got %v", first["id"])
	}
	if first["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", first["name"])
	}
	if first["email"] != "zhangsan@example.com" {
		t.Errorf("Expected email=zhangsan@example.com, got %v", first["email"])
	}
	// 验证未导出的字段不存在
	if _, exists := first["age"]; exists {
		t.Error("Field 'age' should not be exported")
	}
	if _, exists := first["password"]; exists {
		t.Error("Field 'password' should not be exported")
	}

	t.Logf("导出结果:\n%s", string(result))
}

func TestExportModels_CaseInsensitive(t *testing.T) {
	// 测试字段名不区分大小写
	products := []*Product{
		{ID: 1, ProductName: "笔记本电脑", Price: 5999.99, Stock: 10, Category: "电子产品", Description: "高性能笔记本"},
		{ID: 2, ProductName: "鼠标", Price: 99.99, Stock: 50, Category: "外设", Description: "无线鼠标"},
	}

	// 使用不同大小写的字段名
	fields := []string{"ID", "PRODUCT_NAME", "Price"}

	result, err := ExportModels(products, fields)
	if err != nil {
		t.Fatalf("ExportModels failed: %v", err)
	}

	var output []map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// 验证字段存在（应该匹配到小写的 json 标签）
	first := output[0]
	if _, exists := first["id"]; !exists {
		t.Error("Field 'id' should exist")
	}
	if _, exists := first["product_name"]; !exists {
		t.Error("Field 'product_name' should exist")
	}
	if _, exists := first["price"]; !exists {
		t.Error("Field 'price' should exist")
	}

	t.Logf("不区分大小写测试结果:\n%s", string(result))
}

func TestExportModels_WithOmitempty(t *testing.T) {
	// 测试带 omitempty 标签的字段
	orders := []*Order{
		{OrderID: "ORD001", CustomerID: 1001, Amount: 299.99, Status: "paid", CreatedAt: "2024-01-01"},
		{OrderID: "ORD002", CustomerID: 1002, Amount: 0, Status: "pending", CreatedAt: "2024-01-02"},
	}

	// 导出所有字段
	fields := []string{"order_id", "customer_id", "amount", "status", "created_at"}

	result, err := ExportModels(orders, fields)
	if err != nil {
		t.Fatalf("ExportModels failed: %v", err)
	}

	var output []map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if len(output) != 2 {
		t.Errorf("Expected 2 records, got %d", len(output))
	}

	t.Logf("带 omitempty 标签测试结果:\n%s", string(result))
}

func TestExportModels_EmptySlice(t *testing.T) {
	// 测试空切片
	users := []*User{}

	fields := []string{"id", "name"}

	result, err := ExportModels(users, fields)
	if err != nil {
		t.Fatalf("ExportModels failed: %v", err)
	}

	var output []map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if len(output) != 0 {
		t.Errorf("Expected 0 records, got %d", len(output))
	}

	t.Logf("空切片测试结果:\n%s", string(result))
}

func TestExportModels_NilPointer(t *testing.T) {
	// 测试包含 nil 指针的切片
	users := []*User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25},
		nil, // nil 指针应该被跳过
		{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 28},
	}

	fields := []string{"id", "name"}

	result, err := ExportModels(users, fields)
	if err != nil {
		t.Fatalf("ExportModels failed: %v", err)
	}

	var output []map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// 应该只有 2 条记录（nil 指针被跳过）
	if len(output) != 2 {
		t.Errorf("Expected 2 records (nil pointer skipped), got %d", len(output))
	}

	t.Logf("包含 nil 指针测试结果:\n%s", string(result))
}

func TestExportModels_ErrorCases(t *testing.T) {
	// 测试 nil 输入
	_, err := ExportModels(nil, []string{"id"})
	if err == nil {
		t.Error("Expected error for nil models, got nil")
	}

	// 测试 nil fields
	users := []*User{{ID: 1, Name: "张三"}}
	_, err = ExportModels(users, nil)
	if err == nil {
		t.Error("Expected error for nil fields, got nil")
	}

	// 测试非切片类型
	_, err = ExportModels("not a slice", []string{"id"})
	if err == nil {
		t.Error("Expected error for non-slice input, got nil")
	}
}

func TestExportModels_RealWorldExample(t *testing.T) {
	// 真实场景示例：导出用户列表的部分字段用于 API 响应
	users := []*User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25, Password: "secret123"},
		{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 30, Password: "secret456"},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 28, Password: "secret789"},
	}

	// 只导出公开字段，隐藏敏感信息
	publicFields := []string{"id", "name", "email", "age"}

	result, err := ExportModels(users, publicFields)
	if err != nil {
		t.Fatalf("ExportModels failed: %v", err)
	}

	t.Logf("真实场景示例 - 用户公开信息导出:\n%s", string(result))

	// 验证密码字段不存在
	var output []map[string]interface{}
	json.Unmarshal(result, &output)
	for i, user := range output {
		if _, exists := user["password"]; exists {
			t.Errorf("Record %d: password field should not be exported", i)
		}
	}
}
