# Toolbox

Go 工具库集合，提供常用安全与数据处理能力，可直接 `go get` 导入使用。

## 安装

```bash
go get github.com/polite-007/toolbox
```

## 包列表

| 包 | 导入路径 | 说明 |
|---|---|---|
| **fofa** | `github.com/polite-007/toolbox/fofa` | FOFA API 客户端，支持搜索、统计、资产查询 |
| **httpx** | `github.com/polite-007/toolbox/httpx` | HTTP 探测封装（基于 projectdiscovery/httpx） |
| **xlsx** | `github.com/polite-007/toolbox/xlsx` | Excel 读写工具，支持多 Sheet、JSON 导出 |
| **dig** | `github.com/polite-007/toolbox/dig` | DNS 查询工具，支持多记录类型、自定义 DNS 服务器 |
| **goscanner** | `github.com/polite-007/toolbox/goscanner` | 漏洞扫描器封装（基于 goby/goscanner） |

## 使用示例

### FOFA 搜索

```go
import "github.com/polite-007/toolbox/fofa"

client := fofa.NewClient("your@email.com", "your_api_key")
resp, err := client.Search(&fofa.SearchRequest{
    Query:  `title="login"`,
    Size:   100,
    Fields: "ip,port,host,title",
})
```

### HTTP 探测

```go
import "github.com/polite-007/toolbox/httpx"

c := httpx.NewHttpxClient(
    httpx.WithThreads(50),
    httpx.WithTechDetect(),
)
err := c.Run(targets, func(r runner.Result) {
    fmt.Println(r.URL, r.StatusCode)
})
```

### DNS 查询

```go
import "github.com/polite-007/toolbox/dig"

d := dig.NewDig("8.8.8.8:53", 5*time.Second)
result := d.Query("example.com", "A")
fmt.Println(result.Records)
```

### Excel 读写

```go
import "github.com/polite-007/toolbox/xlsx"

data := []xlsx.SheetData{
    {SheetName: "Sheet1", Titles: []string{"Name", "Age"}, Data: [][]string{{"Alice", "30"}}},
}
err := xlsx.ExportToExcel(data, "output.xlsx")
```

## License

[LICENSE](./LICENSE)
