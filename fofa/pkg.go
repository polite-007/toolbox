package fofa

import (
	"reflect"
	"strings"
	"time"
)

// newRecordFromStrings 从字符串数组创建Record，利用反射赋值字段
func newRecordFromStrings(res, fields []string) *SearchResult {
	var r SearchResult
	//  补齐字段，以防万一，防止字段缺失导致panic
	if diff := len(fields) - len(res); diff > 0 {
		emptyStrings := make([]string, diff)
		res = append(res, emptyStrings...)
	}
_nextField:
	for i, v := range fields {
		t := reflect.TypeOf(r)
		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			tag := field.Tag.Get("json")
			if strings.Contains(tag, ",omitempty") {
				tag = strings.Split(tag, ",")[0]
			}
			// 适配小版本，因为FOSS中'product.version'表示 {"product": {"version": "xxx"}}
			if v == "product.version" {
				v = "product_version"
			}
			if v == tag {
				switch tag {
				case "lastupdatetime":
					timeValue, err := time.Parse("2006-01-02 15:04:05", res[i])
					if err != nil {
						timeValue = time.Now()
					}
					reflect.ValueOf(&r).Elem().FieldByName(field.Name).SetString(timeValue.Format("2006-01-02 15:04:05"))
					continue _nextField
				case "longitude", "latitude", "as_number", "icon_hash":
					reflect.ValueOf(&r).Elem().FieldByName(field.Name).SetString(res[i])
					continue _nextField
				default:
					reflect.ValueOf(&r).Elem().FieldByName(field.Name).SetString(res[i])
					continue _nextField
				}
			}
		}
	}
	return &r
}

// setFieldValue 根据字段名设置结构体字段的值
func setFieldValue(result *SearchResult, fieldName, value string) {
	// 使用反射设置字段值
	rv := reflect.ValueOf(result).Elem()
	rt := rv.Type()

	// 遍历所有字段，查找匹配的 json tag
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		jsonTag := field.Tag.Get("json")

		// 检查 json tag 是否匹配字段名
		if jsonTag == fieldName {
			fieldValue := rv.Field(i)
			if fieldValue.CanSet() {
				fieldValue.SetString(value)
			}
			return
		}
	}
}

// getFieldValue 根据字段名从结构体中获取值
func getFieldValue(result *SearchResult, fieldName string) string {
	rv := reflect.ValueOf(result).Elem()
	rt := rv.Type()

	// 遍历所有字段，查找匹配的 json tag
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		jsonTag := field.Tag.Get("json")

		// 检查 json tag 是否匹配字段名
		if jsonTag == fieldName {
			fieldValue := rv.Field(i)
			return fieldValue.String()
		}
	}

	// 如果找不到匹配的字段，返回空字符串
	return ""
}
