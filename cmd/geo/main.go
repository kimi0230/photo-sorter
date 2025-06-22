package main

import (
	"bytes"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// 全局變數
var (
	dbPath   string
	lat, lon float64
	showHelp bool
	cmd      string
)

// init 初始化程序
func init() {
	// 定義命令行參數
	flag.StringVar(&dbPath, "db_path", "states.sqlite", "Spatialite 數據庫路徑")
	flag.Float64Var(&lat, "lat", 43.06417, "緯度")
	flag.Float64Var(&lon, "lon", 141.34694, "經度")
	flag.BoolVar(&showHelp, "help", false, "顯示幫助信息")
	flag.StringVar(&cmd, "cmd", "location", "執行的命令: location(查詢位置) 或 geometry(查詢表結構)")
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

func queryGeometryColumn(dbPath string) (string, error) {
	// 查詢表結構
	sql := `
	SELECT sql FROM sqlite_master WHERE type='table' AND name='ne_10m_admin_1_states_provinces';`
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
	countSQL := `SELECT COUNT(*) FROM ne_10m_admin_1_states_provinces;`
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
	formattedResult := formatTableStructure(result, count)

	return formattedResult, nil
}

// formatTableStructure 格式化表格結構輸出
func formatTableStructure(sql string, count string) string {
	if sql == "" {
		return "未找到表結構信息"
	}

	// 移除多餘的空白和換行
	sql = strings.TrimSpace(sql)

	// 格式化 CREATE TABLE 語句
	formatted := "CREATE TABLE ne_10m_admin_1_states_provinces (\n"

	// 提取欄位定義
	start := strings.Index(sql, "(")
	end := strings.LastIndex(sql, ")")
	if start != -1 && end != -1 && end > start {
		fields := sql[start+1 : end]

		// 分割欄位
		fieldList := strings.Split(fields, ",")
		for i, field := range fieldList {
			field = strings.TrimSpace(field)
			if field != "" {
				// 添加適當的縮排
				formatted += "  " + field
				if i < len(fieldList)-1 {
					formatted += ","
				}
				formatted += "\n"
			}
		}
	}

	formatted += ");"

	// 添加表信息
	info := "\n\n表信息:\n"
	info += "├─ 表名: ne_10m_admin_1_states_provinces\n"
	info += "├─ 類型: 空間表 (Spatial Table)\n"
	info += "├─ 記錄數量: " + count + " 筆\n"
	info += "├─ 描述: 自然地球 1:10M 行政區劃分 (州/省級別)\n"
	info += "└─ 主要欄位:\n"
	info += "   ├─ name: 地區名稱\n"
	info += "   ├─ admin: 行政區名稱\n"
	info += "   ├─ adm0_a3: 國家代碼 (ISO 3166-1 alpha-3)\n"
	info += "   └─ GEOMETRY: 空間幾何數據\n"

	return formatted + info
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
		fmt.Println("  go run cmd/geo/main.go -cmd=location -lat=25.0330 -lon=121.5654")
		fmt.Println("  go run cmd/geo/main.go -lat=35.6895 -lon=139.6917  # 東京")
		fmt.Println("  go run cmd/geo/main.go -lat=40.7128 -lon=-74.0060  # 紐約")
		fmt.Println("")
		fmt.Println("  # 查詢表結構")
		fmt.Println("  go run cmd/geo/main.go -cmd=geometry")
		fmt.Println("  go run cmd/geo/main.go -cmd=geometry -db_path=my_database.sqlite")
		return
	}

	// 根據命令顯示不同的信息
	if cmd == "location" {
		fmt.Printf("查詢位置: 緯度=%.6f, 經度=%.6f\n", lat, lon)
	} else if cmd == "geometry" {
		fmt.Printf("查詢表結構\n")
	}
	fmt.Printf("使用數據庫: %s\n", dbPath)
	fmt.Println("---")

	var result string
	var err error

	if cmd == "location" {
		result, err = queryLocation(dbPath, lon, lat)
	} else if cmd == "geometry" {
		result, err = queryGeometryColumn(dbPath)
	} else {
		err = fmt.Errorf("未知的命令: %s", cmd)
	}

	if err != nil {
		fmt.Printf("錯誤: %v\n", err)
		return
	}

	if result == "" {
		if cmd == "location" {
			fmt.Println("未找到匹配的位置")
		} else {
			fmt.Println("未找到表結構信息")
		}
	} else {
		if cmd == "location" {
			fmt.Println("查詢結果:")
		} else {
			fmt.Println("表結構:")
		}
		fmt.Println(result)
	}
}
