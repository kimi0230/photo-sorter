package main

import (
	"flag"
	"fmt"
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
	switch cmd {
	case CmdLocation:
		handleLocationCommand(dbPath, lat, lon)

	case CmdGeometry:
		handleGeometryCommand(dbPath, tableName)

	default:
		fmt.Printf("錯誤: 未知的命令: %s\n", cmd)
		return
	}
}
