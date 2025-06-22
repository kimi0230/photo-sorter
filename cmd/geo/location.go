package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// queryLocation 查詢地理位置
func queryLocation(dbPath string, lon, lat float64) (string, error) {
	pointWKT := fmt.Sprintf("POINT(%f %f)", lon, lat)

	// 建立 spatialite 查詢語句
	sql := fmt.Sprintf(`
	SELECT name, admin, adm0_a3,name_en FROM ne_10m_admin_1_states_provinces
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

// handleLocationCommand 處理位置查詢命令
func handleLocationCommand(dbPath string, lat, lon float64) {
	fmt.Printf("查詢位置: 緯度=%.6f, 經度=%.6f\n", lat, lon)
	fmt.Printf("使用數據庫: %s\n", dbPath)
	fmt.Println("---")

	result, err := queryLocation(dbPath, lon, lat)
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
}
