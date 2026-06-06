package fileutils_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/polite-007/toolbox/fileutils"
)

// 示例：按行读取文件内容，自动跳过空行
func ExampleReadLines() {
	lines, err := fileutils.ReadLines("fileutils.go")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("total lines:", len(lines))
}

// 示例：将字符串切片逐行写入文件
func ExampleWriteLines() {
	path := filepath.Join(os.TempDir(), "example.txt")

	// 第三个参数 false 表示覆盖写入，true 表示追加
	err := fileutils.WriteLines(path, []string{"hello", "world"}, false)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("written to:", path)
}

// 示例：遍历目录下所有文件，获取文件名列表
func ExampleReadFilesFromDir() {
	// 第二个参数 false 表示只获取文件名，不读取内容
	files, err := fileutils.ReadFilesFromDir(".", false)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, f := range files {
		fmt.Println(f.Name)
	}
}
