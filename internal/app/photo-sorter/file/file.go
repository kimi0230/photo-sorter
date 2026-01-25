package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/geocoding"
	"photo-sorter/internal/pkg/logger"
	"photo-sorter/internal/pkg/metadata"
	"photo-sorter/internal/pkg/tagger"

	"go.uber.org/zap"
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
	exifData, err := metadata.GetExifData(path)
	if err != nil {
		logger.LogInfo(path, zap.String("取得 EXIF 資料失敗", "將檔案移動到失敗資料夾"))
		return HandelFailedFolder(path, cfg, logger)
	}

	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("處理被取消: %v", ctx.Err())
	default:
	}

	// 取得目標路徑
	targetPath, err := metadata.GetTargetPath(path, exifData, cfg)
	if err != nil {
		logger.LogError(path, fmt.Sprintf("取得目標路徑失敗: %v", err))
		return fmt.Errorf("取得目標路徑失敗: %v", err)
	}

	if cfg.DryRun {
		fmt.Printf("DryRun: 將移動: %s -> %s\n", path, targetPath)
		return nil
	}

	// 移動檔案
	if err := MoveFile(path, targetPath); err != nil {
		logger.LogError(path, fmt.Sprintf("移動檔案失敗: %v", err))
		return fmt.Errorf("移動檔案失敗: %v", err)
	}

	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("處理被取消: %v", ctx.Err())
	default:
	}

	// 如果有啟用地理位置標籤且有 GPS 資訊，則為目標檔案添加標籤
	if cfg.EnableGeoTag && exifData.GPSLatitude != "" && exifData.GPSLongitude != "" {
		lat, err := metadata.ParseGPSString(exifData.GPSLatitude)
		if err != nil {
			return fmt.Errorf("解析緯度失敗: %v", err)
		}

		lon, err := metadata.ParseGPSString(exifData.GPSLongitude)
		if err != nil {
			return fmt.Errorf("解析經度失敗: %v", err)
		}

		if lat != 0 && lon != 0 {
			geocoder, err := geocoding.NewGeocoder(cfg.GeocoderType, map[string]interface{}{
				"db_path": cfg.GeoDBPath,
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

	return MoveFile(path, targetPath)
}

// HandelFailedFolder 將檔案移動到失敗資料夾
func HandelFailedFolder(path string, cfg *config.Config, logger *logger.Logger) error {
	// 建立失敗資料夾
	failDir := filepath.Join(cfg.DstDir, "failed_files")

	if err := os.MkdirAll(failDir, 0755); err != nil {
		logger.LogError(path, fmt.Sprintf("建立 failed_files 資料夾失敗: %v", err))
		return err
	}

	// 如果目標檔案已存在，添加時間戳記
	baseName := filepath.Base(path)
	targetPath := filepath.Join(failDir, baseName)
	counter := 1
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		targetPath = filepath.Join(failDir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
		counter++
	}

	if cfg.DryRun {
		fmt.Printf("DryRun: 將移動失敗的檔案: %s -> %s\n", path, targetPath)
		return nil
	}
	return MoveFile(path, targetPath)

}

// MoveFile 移動檔案（保留所有時間戳記）
func MoveFile(src, dst string) error {
	// 確保目標目錄存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("建立目標目錄失敗: %v", err)
	}

	// 嘗試直接移動（同一個檔案系統內）
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// 如果跨檔案系統移動失敗，則複製後刪除
	// 先複製檔案
	if err := copyFileForMove(src, dst); err != nil {
		return fmt.Errorf("複製檔案失敗: %v", err)
	}

	// 刪除原始檔案
	if err := os.Remove(src); err != nil {
		// 如果刪除失敗，嘗試刪除已複製的檔案
		os.Remove(dst)
		return fmt.Errorf("刪除原始檔案失敗: %v", err)
	}

	return nil
}

// copyFileForMove 用於跨檔案系統移動時的複製（保留時間戳記）
func copyFileForMove(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// 取得原始檔案的資訊
	fileInfo, err := source.Stat()
	if err != nil {
		return err
	}
	modTime := fileInfo.ModTime()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 複製內容
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// 保留原始檔案的修改時間和訪問時間
	if err := os.Chtimes(dst, modTime, modTime); err != nil {
		return err
	}

	return nil
}

// CopyFileWithBuffer 使用 buffer 複製檔案（適合大檔案）
func CopyFileWithBuffer(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 取得檔案大小
	fileInfo, err := source.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// 根據檔案大小決定 buffer 大小
	var bufferSize int
	switch {
	case fileSize < 1024*1024: // 小於 1MB
		bufferSize = 32 * 1024 // 32KB
	case fileSize < 10*1024*1024: // 1MB 到 10MB
		bufferSize = 256 * 1024 // 256KB
	case fileSize < 100*1024*1024: // 10MB 到 100MB
		bufferSize = 1024 * 1024 // 1MB
	default: // 大於 100MB
		bufferSize = 4 * 1024 * 1024 // 4MB
	}

	buffer := make([]byte, bufferSize)
	for {
		n, err := source.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buffer[:n]); err != nil {
			return err
		}
	}
	return nil
}

// CopyFileDirect 直接複製檔案（適合小檔案）
func CopyFileDirect(src, dst string) error {
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

// CopyFile 智能選擇的檔案複製函數
func CopyFile(src, dst string) error {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 大檔案使用 Buffer 版本
	if fileInfo.Size() > 100*1024*1024 {
		return CopyFileWithBuffer(src, dst)
	}

	// 小檔案使用直接複製版本
	return CopyFileDirect(src, dst)
}
