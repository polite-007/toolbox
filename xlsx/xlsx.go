/*
 * Copyright (c) 2024. HuaShunXinAn. All rights reserved.
 */
package xlsx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

const (
	DefaultAssetExportSheetName = "Sheet1"
)

// SheetData 表格单页数据
type SheetData struct {
	SheetName string
	Titles    []string
	Data      [][]string
}

func setByRow(f *excelize.File, sheet string, row int, data []string) error {
	// 优化性能：先计算所有单元格地址
	for i, v := range data {
		err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", columnIndexToName(i), row), v)
		if err != nil {
			return errors.WithMessage(err, "SetCellValue failed")
		}
	}
	return nil
}

func generateExcel(f *excelize.File, sheets []SheetData) error {
	for i, sheet := range sheets {
		sheetName := sheet.SheetName
		if sheetName == "" {
			sheetName = DefaultAssetExportSheetName
		}

		// 处理第一个sheet：如果是默认名称则使用，否则重命名默认sheet
		if i == 0 {
			defaultSheetName := f.GetSheetName(0)
			if sheetName != defaultSheetName {
				if err := f.SetSheetName(defaultSheetName, sheetName); err != nil {
					return errors.WithMessage(err, "重命名sheet失败")
				}
			}
		} else {
			// 后续sheet需要创建新sheet
			if _, err := f.NewSheet(sheetName); err != nil {
				return errors.WithMessage(err, "创建sheet失败")
			}
		}

		if err := setByRow(f, sheetName, 1, sheet.Titles); err != nil {
			return errors.WithMessage(err, "setByRow failed")
		}
		for j, v := range sheet.Data {
			if err := setByRow(f, sheetName, j+2, v); err != nil {
				return errors.WithMessage(err, "setByRow failed")
			}
		}
	}
	return nil
}

func checkAndCreateFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return errors.WithMessage(err, "创建目录失败")
		}
		// 文件不存在，创建文件
		file, err := os.Create(filePath)
		if err != nil {
			return errors.WithMessage(err, "创建文件失败")
		}
		// 关闭文件
		defer file.Close()
		log.Printf("文件 %s 已创建", filePath)
	} else if err != nil {
		return errors.WithMessage(err, "检查文件失败")
	} else {
		log.Printf("文件 %s 已存在", filePath)
	}
	return nil
}

func columnIndexToName(index int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if index < 0 {
		return ""
	}
	if index < len(letters) {
		return string(letters[index])
	}

	var name strings.Builder
	for index > 0 {
		index--
		remainder := index % len(letters)
		name.WriteByte(letters[remainder])
		index /= len(letters)
	}

	// Reverse the string to get the correct column name
	runes := []rune(name.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// GenerateExcelToFile 将一个或多个工作表数据写入到Excel文件中
//
// 参数：
// fileName: 输出的Excel文件名
// sheets: 包含要写入的工作表数据的切片
//
// 返回值：
// error: 如果操作过程中出现错误，则返回错误信息；否则返回nil
func GenerateExcelToFile(fileName string, sheets []SheetData) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	if err := checkAndCreateFile(fileName); err != nil {
		return errors.WithMessage(err, "checkAndCreateFile failed")
	}

	if err := generateExcel(f, sheets); err != nil {
		return errors.WithMessage(err, "generateExcel failed")
	}

	if err := f.SaveAs(fileName); err != nil {
		return errors.WithMessage(err, "saveAs failed")
	}

	return nil
}
