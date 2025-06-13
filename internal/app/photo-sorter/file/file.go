package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/exif"
	"photo-sorter/internal/pkg/geocoding"
	"photo-sorter/internal/pkg/logger"
	"photo-sorter/internal/pkg/tagger"
)

// ProcessFile 處理單個檔案
func ProcessFile(ctx context.Context, path string, cfg *config.Config, logger *logger.Logger) error {
	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("處理被取消: %v", ctx.Err())
	default:
	}

	// 取得 EXIF 資料
	exifData, err := exif.GetExifData(path)
	if err != nil {
		logger.LogError(path, fmt.Sprintf("取得 EXIF 資料失敗: %v", err))
		return fmt.Errorf("取得 EXIF 資料失敗: %v", err)
	}

	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("處理被取消: %v", ctx.Err())
	default:
	}

	// 取得目標路徑
	targetPath, err := exif.GetTargetPath(path, exifData, cfg)
	if err != nil {
		logger.LogError(path, fmt.Sprintf("取得目標路徑失敗: %v", err))
		return fmt.Errorf("取得目標路徑失敗: %v", err)
	}

	if cfg.DryRun {
		fmt.Printf("DryRun: 將移動: %s -> %s\n", path, targetPath)
		return nil
	}

	// 複製檔案
	if err := CopyFile(path, targetPath); err != nil {
		logger.LogError(path, fmt.Sprintf("複製檔案失敗: %v", err))
		return fmt.Errorf("複製檔案失敗: %v", err)
	}

	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("處理被取消: %v", ctx.Err())
	default:
	}

	// 如果有啟用地理位置標籤且有 GPS 資訊，則為目標檔案添加標籤
	if cfg.EnableGeoTag && exifData.GPSLatitude != "" && exifData.GPSLongitude != "" {
		lat, err := exif.ParseGPSString(exifData.GPSLatitude)
		if err != nil {
			return fmt.Errorf("解析緯度失敗: %v", err)
		}

		lon, err := exif.ParseGPSString(exifData.GPSLongitude)
		if err != nil {
			return fmt.Errorf("解析經度失敗: %v", err)
		}

		if lat != 0 && lon != 0 {
			geocoder, err := geocoding.NewGeocoder(cfg.GeocoderType, map[string]interface{}{
				"json_path": cfg.GeoJSONPath,
			})
			if err == nil {
				countryCity, err := geocoder.GetLocationFromGPS(lat, lon)
				if err == nil && countryCity != nil {
					if !cfg.DryRun {
						fileTagger, err := tagger.NewTagger()
						if err != nil {
							return fmt.Errorf("建立標籤實例失敗: %v", err)
						}
						tagName := fmt.Sprintf("%s-%s", countryCity.Country, strings.ReplaceAll(countryCity.City, " ", "_"))
						if err := fileTagger.AddTag(targetPath, tagName); err != nil {
							fmt.Printf("為檔案添加標籤失敗: %v\n", err)
						}
					} else {
						fmt.Printf("DryRun: 為檔案添加標籤: %s\n", targetPath)
					}
				}
			}
		}
	}

	return nil
}

// CopyFile 複製檔案
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

// HandleUnsupportedFile 處理不支援的檔案
func HandleUnsupportedFile(path string, cfg *config.Config, logger *logger.Logger) error {
	// 建立 unknown_format 資料夾
	unknownDir := filepath.Join(cfg.DstDir, "unknown_format")
	if err := os.MkdirAll(unknownDir, 0755); err != nil {
		logger.LogError(path, fmt.Sprintf("建立 unknown_format 資料夾失敗: %v", err))
		return err
	}

	// 處理檔案名稱衝突
	baseName := filepath.Base(path)
	targetPath := filepath.Join(unknownDir, baseName)
	counter := 1
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		targetPath = filepath.Join(unknownDir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
		counter++
	}

	if cfg.DryRun {
		fmt.Printf("DryRun: 將移動不支援的檔案: %s -> %s\n", path, targetPath)
		return nil
	}

	return CopyFile(path, targetPath)
}
