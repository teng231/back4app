package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Hàm tính độ dài tối đa của mỗi cột
func getMaxLengths(data []map[string]interface{}) map[string]int {
	maxLengths := make(map[string]int)
	for _, row := range data {
		for key, value := range row {
			valueStr := fmt.Sprintf("%v", value)
			if len(key) > maxLengths[key] {
				maxLengths[key] = len(key)
			}
			if len(valueStr) > maxLengths[key] {
				maxLengths[key] = len(valueStr)
			}
		}
	}
	return maxLengths
}

// Hàm định dạng chuỗi đầu ra dạng bảng
func PrintTable(data []map[string]interface{}) string {
	if len(data) == 0 {
		return ""
	}

	headers := make([]string, 0)
	for key := range data[0] {
		headers = append(headers, key)
	}
	sort.Strings(headers) // Sắp xếp headers theo thứ tự bảng chữ cái
	maxLengths := getMaxLengths(data)

	// Định dạng header
	headerRow := ""
	separatorRow := ""
	for _, header := range headers {
		headerRow += fmt.Sprintf("%-*s|", maxLengths[header], header)
		separatorRow += strings.Repeat("-", maxLengths[header]) + "+"
	}

	result := headerRow + "\n" + separatorRow + "\n"

	// Định dạng từng dòng dữ liệu
	for _, row := range data {
		rowStr := ""
		for _, header := range headers {
			value := row[header]
			valueStr := fmt.Sprintf("%v", value)
			rowStr += fmt.Sprintf("%-*s|", maxLengths[header], valueStr)
		}
		result += rowStr + "\n"
	}

	return result
}

// Chuyển struct slice thành map slice

func StructSliceToMapSlice(slice interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	val := reflect.ValueOf(slice)
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()

		// Kiểm tra nếu là con trỏ đến struct, lấy giá trị của struct
		if reflect.TypeOf(item).Kind() == reflect.Ptr {
			item = reflect.ValueOf(item).Elem().Interface()
		}

		itemMap := make(map[string]interface{})
		itemVal := reflect.ValueOf(item)
		for j := 0; j < itemVal.NumField(); j++ {
			field := itemVal.Type().Field(j)
			itemMap[field.Name] = itemVal.Field(j).Interface()
		}
		result = append(result, itemMap)
	}

	return result
}
