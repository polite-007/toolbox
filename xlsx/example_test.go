package xlsx_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/polite-007/toolbox/xlsx"
)

// 示例：生成 Excel 文件，支持多 Sheet、表头和数据行
func ExampleGenerateExcelToFile() {
	// 定义表格数据：每个 SheetData 表示一个工作表
	sheets := []*xlsx.SheetData{
		{
			SheetName: "Users",                                        // 工作表名称
			Titles:    []string{"Name", "Age", "Email"},               // 表头
			Data: [][]string{
				{"Alice", "30", "alice@example.com"},                  // 数据行
				{"Bob", "25", "bob@example.com"},
			},
		},
	}

	// 导出到文件，目录不存在时会自动创建
	out := filepath.Join(os.TempDir(), "example.xlsx")
	if err := xlsx.GenerateExcelToFile(out, sheets); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("Excel file created at:", out)
}
