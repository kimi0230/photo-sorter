package exif

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/geocoding"
	"photo-sorter/internal/pkg/tagger"
)

type ExifData struct {
	CreateDate      string `json:"CreateDate"`
	MediaCreateDate string `json:"MediaCreateDate"`
	Model           string `json:"Model"`
	GPSLatitude     string `json:"GPSLatitude"`
	GPSLongitude    string `json:"GPSLongitude"`
}

// parseGPSString 將 GPS 字串轉換為浮點數
// 格式範例: "22 deg 41' 58.80\" N"
func parseGPSString(gpsStr string) (float64, error) {
	if gpsStr == "" {
		return 0, nil
	}

	// 移除引號和空格
	gpsStr = strings.Trim(gpsStr, "\"")
	gpsStr = strings.TrimSpace(gpsStr)

	// 分割字串
	parts := strings.Fields(gpsStr)
	if len(parts) < 4 {
		return 0, fmt.Errorf("無效的 GPS 格式: %s", gpsStr)
	}

	// 解析度數
	degrees, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("解析度數失敗: %v", err)
	}

	// 解析分數
	minutes, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "'"), 64)
	if err != nil {
		return 0, fmt.Errorf("解析分數失敗: %v", err)
	}

	// 解析秒數
	seconds, err := strconv.ParseFloat(strings.TrimSuffix(parts[3], "\""), 64)
	if err != nil {
		return 0, fmt.Errorf("解析秒數失敗: %v", err)
	}

	// 計算十進位度數
	decimal := degrees + minutes/60 + seconds/3600

	// 檢查方向（N/S 或 E/W）
	if len(parts) > 4 {
		direction := parts[4]
		if direction == "S" || direction == "W" {
			decimal = -decimal
		}
	}

	return decimal, nil
}

func GetExifData(path string) (*ExifData, error) {
	cmd := exec.Command("exiftool", "-json", "-CreateDate", "-MediaCreateDate", "-Model", "-GPSLatitude", "-GPSLongitude", path)
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

	// 如果有啟用地理位置標籤且有 GPS 資訊，則加入地理位置
	if cfg.EnableGeoTag && exif.GPSLatitude != "" && exif.GPSLongitude != "" {
		lat, err := parseGPSString(exif.GPSLatitude)
		if err != nil {
			return "", fmt.Errorf("解析緯度失敗: %v", err)
		}

		lon, err := parseGPSString(exif.GPSLongitude)
		if err != nil {
			return "", fmt.Errorf("解析經度失敗: %v", err)
		}

		if lat != 0 && lon != 0 {
			geocoder, err := geocoding.NewGeocoder(cfg.GeocoderType, map[string]interface{}{
				"json_path": cfg.GeoJSONPath,
			})
			if err == nil {
				countryCity, err := geocoder.GetLocationFromGPS(lat, lon)
				if err == nil && countryCity != nil {
					// 為檔案添加標籤
					fileTagger, err := tagger.NewTagger()
					if err != nil {
						return "", fmt.Errorf("建立標籤實例失敗: %v", err)
					}
					tagName := fmt.Sprintf("%s-%s", countryCity.Country, strings.ReplaceAll(countryCity.City, " ", "_"))
					if err := fileTagger.AddTag(path, tagName); err != nil {
						fmt.Printf("為檔案添加標籤失敗: %v\n", err)
					}
					date = fmt.Sprintf("%s-%s-%s", date, countryCity.Country, strings.ReplaceAll(countryCity.City, " ", "_"))
				}
			}
		}
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
