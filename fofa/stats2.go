package fofa

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// StatsRequest 统计请求参数
type Stats2Request struct {
	Query    string // 搜索查询语句
	Fields   string // 统计字段，如: protocol,port,country
	MaxCount int    // 最多获取多少个值，默认5, 最大10
}

// Stats2Response 统计响应结果
type Stats2Response struct {
	Content []string
}

// Stats2 使用search加改造fofa语句的逻辑来获取某个字段的聚合结果,
// 注意这个返回的字段不是size为倒序的，是随机的以更新时间为倒序的，且无法获取每个内容的具体数量
// 旨在通过消耗search额度来部分实现stats功能
func (c *Client) Stats2(req *Stats2Request) (*Stats2Response, error) {
	if req.Query == "" {
		return nil, errors.New("查询语句不能为空")
	}

	if req.Fields == "" {
		return nil, errors.New("统计字段不能为空")
	}

	// 使用指定的字段
	targetField := strings.TrimSpace(req.Fields)

	var allValues []string

	var currentQuery string

	switch targetField {
	case "protocol":
		currentQuery = fmt.Sprintf(`type=service && %s `, req.Query)
	default:
		currentQuery = req.Query
	}

	seenValues := make(map[string]bool) // 用于去重

	// 循环查询，直到没有结果
	for {
		// 使用 Search 查询，size=1，只获取一个结果
		searchResp, err := c.Search(&SearchRequest{
			Query:  currentQuery,
			Fields: fmt.Sprintf("ip,%s", req.Fields),
			Size:   1,
			Page:   1,
		})
		if err != nil {
			return nil, errors.Wrap(err, "查询失败")
		}

		// 检查是否有错误
		if searchResp.Error {
			return nil, fmt.Errorf("FOFA API 错误: %s", searchResp.ErrMsg)
		}

		// 如果没有结果，退出循环
		if len(searchResp.Results) == 0 || searchResp.Size == 0 {
			break
		}

		// 获取字段值
		// 找到目标字段在 Fields 中的索引
		fieldIndex := -1
		for i, field := range searchResp.Fields {
			if field == targetField {
				fieldIndex = i
				break
			}
		}

		if fieldIndex == -1 {
			return nil, fmt.Errorf("字段 %s 不在返回结果中", targetField)
		}

		// 从结果中提取字段值
		if len(searchResp.Results) > 0 && len(searchResp.Results[0]) > fieldIndex {
			fieldValue := searchResp.Results[0][fieldIndex]

			// 如果值为空，跳过
			if fieldValue == "" {
				break
			}

			// 去重检查
			if seenValues[fieldValue] {
				// 如果已经见过这个值，说明可能进入了循环，退出
				break
			}
			seenValues[fieldValue] = true
			allValues = append(allValues, fieldValue)

			// 修改查询语句，添加排除条件
			// 转义字段值中的特殊字符（如引号）
			escapedValue := strings.ReplaceAll(fieldValue, `"`, `\"`)
			// 使用 && 连接排除条件
			currentQuery = fmt.Sprintf(`%s && %s!="%s"`, currentQuery, targetField, escapedValue)
		} else {
			// 如果无法获取字段值，退出循环
			break
		}
	}

	return &Stats2Response{
		Content: allValues,
	}, nil
}
