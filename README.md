# Toolbox

Go 工具库集合，提供常用安全与数据处理能力，可直接 `go get` 导入使用。

旨在将很多复杂，好用的网络安全工具，pkg化

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
## License

[LICENSE](./LICENSE)
