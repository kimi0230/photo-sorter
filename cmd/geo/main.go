package main

import (
	"bytes"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// 命令常量
const (
	CmdLocation = "location"
	CmdGeometry = "geometry"
)

// 全局變數
var (
	dbPath    string
	lat, lon  float64
	showHelp  bool
	cmd       string
	tableName string
)

// init 初始化程序
func init() {
	// 定義命令行參數
	flag.StringVar(&dbPath, "db_path", "states.sqlite", "Spatialite 數據庫路徑")
	flag.Float64Var(&lat, "lat", 43.06417, "緯度")
	flag.Float64Var(&lon, "lon", 141.34694, "經度")
	flag.BoolVar(&showHelp, "help", false, "顯示幫助信息")
	flag.StringVar(&cmd, "cmd", CmdLocation, "執行的命令: location(查詢位置) 或 geometry(查詢表結構)")
	flag.StringVar(&tableName, "table_name", "ne_10m_admin_1_states_provinces", "要查詢的表名")
}

// queryLocation 查詢地理位置
func queryLocation(dbPath string, lon, lat float64) (string, error) {
	pointWKT := fmt.Sprintf("POINT(%f %f)", lon, lat)

	// 建立 spatialite 查詢語句
	sql := fmt.Sprintf(`
	SELECT name, admin, adm0_a3 FROM ne_10m_admin_1_states_provinces
	WHERE ST_Contains(
		GEOMETRY,
		ST_PointFromText('%s', 4326)
	);
	`, pointWKT)

	// 使用 tools/bin 下的 spatialite
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
	return result, nil
}

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
	fieldCount := 0

	for _, field := range fieldList {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		// 提取欄位名和類型
		parts := strings.Fields(field)
		if len(parts) >= 2 {
			fieldName := parts[0]
			fieldType := parts[1]

			// 根據欄位名和類型提供描述
			description := getFieldDescription(fieldName, fieldType)

			if fieldCount < len(fieldList)-2 { // 不是最後一個欄位
				fieldInfo += fmt.Sprintf("   ├─ %s (%s): %s\n", fieldName, fieldType, description)
			} else {
				fieldInfo += fmt.Sprintf("   └─ %s (%s): %s\n", fieldName, fieldType, description)
			}
			fieldCount++
		}
	}

	if fieldInfo == "" {
		fieldInfo = "   └─ 無法解析欄位信息\n"
	}

	return fieldInfo
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

func main() {
	// 解析命令行參數
	flag.Parse()

	// 顯示幫助信息
	if showHelp {
		fmt.Println("地理編碼查詢工具")
		fmt.Println("用法:")
		fmt.Println("  go run cmd/geo/main.go [選項]")
		fmt.Println("")
		fmt.Println("選項:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("命令範例:")
		fmt.Println("  # 查詢位置 (預設)")
		fmt.Printf("  go run cmd/geo/main.go -cmd=%s -lat=25.0330 -lon=121.5654\n", CmdLocation)
		fmt.Println("  go run cmd/geo/main.go -lat=35.6895 -lon=139.6917  # 東京")
		fmt.Println("  go run cmd/geo/main.go -lat=40.7128 -lon=-74.0060  # 紐約")
		fmt.Println("")
		fmt.Println("  # 查詢表結構")
		fmt.Printf("  go run cmd/geo/main.go -cmd=%s\n", CmdGeometry)
		fmt.Printf("  go run cmd/geo/main.go -cmd=%s -table_name=my_table\n", CmdGeometry)
		fmt.Printf("  go run cmd/geo/main.go -cmd=%s -db_path=my_database.sqlite -table_name=spatial_table\n", CmdGeometry)
		return
	}

	// 根據命令執行相應操作
	var result string
	var err error

	switch cmd {
	case CmdLocation:
		fmt.Printf("查詢位置: 緯度=%.6f, 經度=%.6f\n", lat, lon)
		fmt.Printf("使用數據庫: %s\n", dbPath)
		fmt.Println("---")
		result, err = queryLocation(dbPath, lon, lat)
		if err != nil {
			fmt.Printf("錯誤: %v\n", err)
			return
		}
		if result == "" {
			fmt.Println("未找到匹配的位置")
		} else {
			fmt.Println("查詢結果:")
			fmt.Println(result)
		}

	case CmdGeometry:
		fmt.Printf("查詢表結構\n")
		fmt.Printf("使用數據庫: %s\n", dbPath)
		fmt.Println("---")
		result, err = queryGeometryColumn(dbPath, tableName)
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

	default:
		err = fmt.Errorf("未知的命令: %s", cmd)
		fmt.Printf("錯誤: %v\n", err)
		return
	}
}
