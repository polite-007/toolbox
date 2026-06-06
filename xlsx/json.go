package xlsx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// exportModels 通用结构体导出工具
// 参数说明：
// - models: 任意结构体指针切片（必须是 []*Struct 类型）
// - fields: 需要导出的字段列表（匹配json标签，不区分大小写）
func ExportModels(models interface{}, fields []string) ([]byte, error) {
	// 输入校验
	if models == nil || fields == nil {
		return nil, fmt.Errorf("models and fields cannot be nil")
	}

	val := reflect.ValueOf(models)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("models must be a slice")
	}

	// 准备字段集合（小写匹配）
	fieldSet := make(map[string]struct{})
	for _, f := range fields {
		fieldSet[strings.ToLower(f)] = struct{}{}
	}

	result := make([]map[string]interface{}, 0, val.Len())

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() != reflect.Ptr || elem.IsNil() {
			continue // 跳过非指针或空指针
		}

		structVal := elem.Elem()
		if structVal.Kind() != reflect.Struct {
			continue // 跳过非结构体
		}

		selected := make(map[string]interface{})
		structType := structVal.Type()

		for j := 0; j < structVal.NumField(); j++ {
			field := structType.Field(j)

			// 解析json标签（兼容带选项的情况，如 `json:"name,omitempty"`）
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = field.Name
			} else {
				jsonTag = strings.Split(jsonTag, ",")[0] // 取第一个部分
			}

			// 字段匹配（不区分大小写）
			if _, exists := fieldSet[strings.ToLower(jsonTag)]; exists {
				fieldValue := structVal.Field(j).Interface()
				selected[jsonTag] = fieldValue
			}
		}

		if len(selected) > 0 {
			result = append(result, selected)
		}
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return jsonData, nil
}
