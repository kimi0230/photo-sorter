package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// queryGeometryColumn 查詢表結構
func queryGeometryColumn(dbPath, tableName string) (string, error) {
	// 查詢表結構
	sql := fmt.Sprintf(`
	SELECT sql FROM sqlite_master WHERE type='table' AND name='%s';`, tableName)
	spatialitePath := filepath.Join("spatialite")
	cmd := exec.Command(spatialitePath, dbPath)

	// 傳入 SQL 並執行
	cmd.Stdin = strings.NewReader(sql)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("spatialite command error: %v, output: %s", err, out.String())
	}

	result := strings.TrimSpace(out.String())

	// 查詢記錄數量
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, tableName)
	cmd = exec.Command(spatialitePath, dbPath)
	cmd.Stdin = strings.NewReader(countSQL)
	var countOut bytes.Buffer
	cmd.Stdout = &countOut
	cmd.Stderr = &countOut

	err = cmd.Run()
	count := "未知"
	if err == nil {
		count = strings.TrimSpace(countOut.String())
	}

	// 格式化輸出
	formattedResult := formatTableStructure(result, count, tableName)

	return formattedResult, nil
}

// handleGeometryCommand 處理表結構查詢命令
func handleGeometryCommand(dbPath, tableName string) {
	fmt.Printf("查詢表結構\n")
	fmt.Printf("使用數據庫: %s\n", dbPath)
	fmt.Println("---")

	result, err := queryGeometryColumn(dbPath, tableName)
	if err != nil {
		fmt.Printf("錯誤: %v\n", err)
		return
	}

	if result == "" {
		fmt.Println("未找到表結構信息")
	} else {
		fmt.Println("表結構:")
		fmt.Println(result)
	}
}

// formatTableStructure 格式化表格結構輸出
func formatTableStructure(sql string, count string, tableName string) string {
	if sql == "" {
		return fmt.Sprintf("未找到表 '%s' 的結構信息", tableName)
	}

	// 添加表信息
	info := "表信息:\n"
	info += fmt.Sprintf("├─ 表名: %s\n", tableName)
	info += "├─ 類型: 空間表 (Spatial Table)\n"
	info += fmt.Sprintf("├─ 記錄數量: %s 筆\n", count)

	// 根據表名提供不同的描述
	description := getTableDescription(tableName)
	info += fmt.Sprintf("├─ 描述: %s\n", description)

	// 動態分析欄位
	fieldInfo := analyzeTableFields(sql)
	info += "└─ 主要欄位:\n"
	info += fieldInfo

	return info
}

// getTableDescription 根據表名返回描述
func getTableDescription(tableName string) string {
	switch tableName {
	case "ne_10m_admin_1_states_provinces":
		return "自然地球 1:10M 行政區劃分 (州/省級別)"
	case "ne_10m_admin_0_countries":
		return "自然地球 1:10M 國家邊界"
	case "ne_10m_populated_places":
		return "自然地球 1:10M 人口聚集地"
	default:
		return "空間數據表"
	}
}

// analyzeTableFields 分析表欄位並返回描述
func analyzeTableFields(sql string) string {
	// 提取欄位定義
	start := strings.Index(sql, "(")
	end := strings.LastIndex(sql, ")")
	if start == -1 || end == -1 || end <= start {
		return "   └─ 無法解析欄位信息\n"
	}

	fields := sql[start+1 : end]
	fieldList := strings.Split(fields, ",")

	var fieldInfo string
	validFields := 0

	for _, field := range fieldList {
		field = strings.TrimSpace(field)
		if field == "" || strings.Contains(strings.ToUpper(field), "PRIMARY KEY") {
			continue
		}

		// 處理帶引號的欄位名
		fieldName := extractFieldName(field)
		fieldType := extractFieldType(field)

		if fieldName != "" && fieldType != "" {
			description := getFieldDescription(fieldName, fieldType)

			if validFields == 0 {
				fieldInfo += fmt.Sprintf("   ├─ %s (%s): %s\n", fieldName, fieldType, description)
			} else {
				fieldInfo += fmt.Sprintf("   ├─ %s (%s): %s\n", fieldName, fieldType, description)
			}
			validFields++
		}
	}

	if fieldInfo == "" {
		fieldInfo = "   └─ 無法解析欄位信息\n"
	} else {
		// 將最後一個 ├─ 改為 └─
		fieldInfo = strings.TrimSuffix(fieldInfo, "├─ ")
		fieldInfo = strings.TrimSuffix(fieldInfo, "\n")
		fieldInfo += "└─ "
		// 重新添加最後一個欄位的信息
		lastField := fieldList[len(fieldList)-1]
		lastField = strings.TrimSpace(lastField)
		if lastField != "" {
			fieldName := extractFieldName(lastField)
			fieldType := extractFieldType(lastField)
			if fieldName != "" && fieldType != "" {
				description := getFieldDescription(fieldName, fieldType)
				fieldInfo += fmt.Sprintf("%s (%s): %s\n", fieldName, fieldType, description)
			}
		}
	}

	return fieldInfo
}

// extractFieldName 提取欄位名稱
func extractFieldName(field string) string {
	field = strings.TrimSpace(field)

	// 處理帶引號的欄位名
	if strings.HasPrefix(field, `"`) {
		end := strings.Index(field[1:], `"`)
		if end != -1 {
			return field[1 : end+1]
		}
	} else if strings.HasPrefix(field, `'`) {
		end := strings.Index(field[1:], `'`)
		if end != -1 {
			return field[1 : end+1]
		}
	} else {
		// 沒有引號，取第一個空格前的部分
		parts := strings.Fields(field)
		if len(parts) > 0 {
			return parts[0]
		}
	}

	return ""
}

// extractFieldType 提取欄位類型
func extractFieldType(field string) string {
	field = strings.TrimSpace(field)

	// 跳過欄位名，找到類型
	parts := strings.Fields(field)
	if len(parts) < 2 {
		return ""
	}

	// 如果第一個部分是帶引號的欄位名，從第二個部分開始找類型
	fieldType := parts[1]

	// 處理類型後面的約束條件
	if strings.Contains(fieldType, "(") {
		// 類型有長度限制，如 VARCHAR(24)
		return fieldType
	}

	return fieldType
}

// getFieldDescription 根據欄位名和類型返回描述
func getFieldDescription(fieldName, fieldType string) string {
	switch strings.ToLower(fieldName) {
	case "id":
		return "主鍵識別碼"
	case "name":
		return "名稱"
	case "admin":
		return "行政區名稱"
	case "adm0_a3":
		return "國家代碼 (ISO 3166-1 alpha-3)"
	case "geometry":
		return "空間幾何數據"
	case "geom":
		return "空間幾何數據"
	case "the_geom":
		return "空間幾何數據"
	case "lat":
		return "緯度"
	case "lon":
		return "經度"
	case "longitude":
		return "經度"
	case "latitude":
		return "緯度"
	case "population":
		return "人口數量"
	case "area":
		return "面積"
	default:
		// 根據類型提供一般描述
		switch strings.ToUpper(fieldType) {
		case "TEXT", "VARCHAR", "CHAR":
			return "文字資料"
		case "INTEGER", "INT":
			return "整數資料"
		case "REAL", "FLOAT", "DOUBLE":
			return "浮點數資料"
		case "BLOB":
			return "二進制資料"
		case "BOOLEAN", "BOOL":
			return "布林值"
		default:
			return "資料欄位"
		}
	}
}
