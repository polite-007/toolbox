package xlsx

import (
	"testing"
)

func TestXlsx(t *testing.T) {
	_ = SheetData{
		SheetName: "httpx",
		Titles:    []string{"ip", "port", "host"},
		Data:      [][]string{{"127.0.0.1", "80", "www.baidu.com"}, {"127.0.0.1", "443", "www.baidu.com"}},
	}
}
