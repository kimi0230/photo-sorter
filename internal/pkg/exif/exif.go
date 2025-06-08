package exif

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"photo-sorter/internal/pkg/config"
)

type ExifData struct {
	CreateDate      string `json:"CreateDate"`
	MediaCreateDate string `json:"MediaCreateDate"`
	Model           string `json:"Model"`
}

func GetExifData(path string) (*ExifData, error) {
	cmd := exec.Command("exiftool", "-json", "-CreateDate", "-MediaCreateDate", "-Model", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("執行 exiftool 失敗: %v", err)
	}

	var data []ExifData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("解析 exiftool 輸出失敗: %v", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("無法取得檔案資訊")
	}

	return &data[0], nil
}

func GetTargetPath(path string, exif *ExifData, cfg *config.Config) (string, error) {
	// 取得日期
	date := exif.CreateDate
	if date == "" {
		date = exif.MediaCreateDate
	}
	if date == "" {
		date = "unknown_date"
	} else {
		// 解析日期字串並使用設定檔中的格式
		t, err := time.Parse("2006:01:02 15:04:05", date)
		if err != nil {
			return "", fmt.Errorf("解析日期失敗: %v", err)
		}
		date = t.Format(cfg.DateFormat)
	}

	// 取得裝置名稱
	device := exif.Model
	if device == "" {
		device = "unknown_device"
	} else {
		// 處理裝置名稱
		device = strings.ReplaceAll(device, " ", "_")
		device = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				return r
			}
			return -1
		}, device)
	}

	// 建立目標路徑
	targetDir := filepath.Join(cfg.DstDir, date, device)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("建立目標資料夾失敗: %v", err)
	}

	// 處理檔案名稱衝突
	baseName := filepath.Base(path)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	targetPath := filepath.Join(targetDir, baseName)

	counter := 1
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		targetPath = filepath.Join(targetDir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
		counter++
	}

	return targetPath, nil
}
