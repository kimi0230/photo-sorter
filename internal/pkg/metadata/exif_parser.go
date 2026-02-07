package metadata

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"photo-sorter/internal/pkg/config"
)

var exifDateFormats = []string{
	"2006:01:02 15:04:05",       // Standard EXIF format.
	"2006:01:02 15:04:05-07:00", // EXIF with timezone.
	"2006:01:02 15:04:05+07:00", // EXIF with timezone.
	"2006-01-02 15:04:05",       // ISO format.
	"2006-01-02 15:04:05-07:00", // ISO with timezone.
	"2006-01-02 15:04:05+07:00", // ISO with timezone.
}

// ParseGPSString 將 GPS 字串轉換為浮點數
// 格式範例: "22 deg 41' 58.80\" N"
func ParseGPSString(gpsStr string) (float64, error) {
	if gpsStr == "" {
		return 0, nil
	}

	// 移除引號和空格
	gpsStr = strings.Trim(gpsStr, "\"")
	gpsStr = strings.TrimSpace(gpsStr)

	// 分割字串
	parts := strings.Fields(gpsStr)
	if len(parts) < 4 {
		return 0, fmt.Errorf("invalid GPS format: %s", gpsStr)
	}

	// 解析度數
	degrees, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse degrees: %v", err)
	}

	// 解析分數
	minutes, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "'"), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse minutes: %v", err)
	}

	// 解析秒數
	seconds, err := strconv.ParseFloat(strings.TrimSuffix(parts[3], "\""), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse seconds: %v", err)
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

// isValidDate 檢查日期是否有效
// 將空字串和無效日期（如 0000:00:00 00:00:00）視為無效
func isValidDate(date string) bool {
	if date == "" {
		return false
	}
	// 檢查是否為無效日期格式
	invalidDates := []string{
		"0000:00:00 00:00:00",
		"0000-00-00 00:00:00",
		"0000:00:00",
		"0000-00-00",
	}
	for _, invalid := range invalidDates {
		if date == invalid {
			return false
		}
	}
	return true
}

type defaultDateResolver struct{}

func (r *defaultDateResolver) Resolve(exif *ExifData, cfg *config.Config) (string, error) {
	// 取得日期（優先順序：CreateDate > DateTimeCreated > MediaCreateDate > FileModifyDate）
	date := ""
	if isValidDate(exif.CreateDate) {
		date = exif.CreateDate
	} else if isValidDate(exif.DateTimeCreated) {
		date = exif.DateTimeCreated
	} else if isValidDate(exif.MediaCreateDate) {
		date = exif.MediaCreateDate
	} else if isValidDate(exif.FileModifyDate) {
		date = exif.FileModifyDate
	}

	if date == "" {
		return "unknown_date", nil
	}

	// 解析日期字串並使用設定檔中的格式
	// 嘗試多種日期格式
	var t time.Time
	var err error
	for _, format := range exifDateFormats {
		t, err = time.Parse(format, date)
		if err == nil {
			return t.Format(cfg.DateFormat), nil
		}
	}
	return "unknown_date", nil
}

type defaultDeviceResolver struct{}

func (r *defaultDeviceResolver) Resolve(exif *ExifData) string {
	// 取得裝置名稱
	// 優先使用 Model，如果沒有 Model 則使用 Encoder，最後才檢查 Description 是否為 "Screenshot"
	var device string
	if exif.Model != "" {
		device = getDeviceName(exif.Model)
	} else if exif.Encoder != "" {
		device = getDeviceName(exif.Encoder)
	} else {
		device = "unknown_device"
	}
	// 如果沒有裝置名稱，且 Description 是 "Screenshot"，則使用 "Screenshot"
	if device == "unknown_device" && strings.TrimSpace(exif.Description) == "Screenshot" {
		device = "Screenshot"
	}
	return device
}

// ResolveDate resolves a normalized date string using default rules.
func ResolveDate(exif *ExifData, cfg *config.Config) (string, error) {
	return (&defaultDateResolver{}).Resolve(exif, cfg)
}

// ResolveDevice resolves a normalized device name using default rules.
func ResolveDevice(exif *ExifData) string {
	return (&defaultDeviceResolver{}).Resolve(exif)
}

// sanitizeDeviceName 清理並標準化裝置名稱
func sanitizeDeviceName(name string) string {
	// 移除所有非英文字母、數字和底線的字元
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, name)
}

// getDeviceName 取得標準化的裝置名稱
func getDeviceName(model string) string {
	if model == "" {
		return "unknown_device"
	}

	// 處理 DJI 機型名稱
	if deviceDJI, ok := GetDJIModelFriendlyName(model); ok {
		return deviceDJI
	}

	// 處理一般裝置名稱
	device := strings.ReplaceAll(model, " ", "_")
	return sanitizeDeviceName(device)
}
