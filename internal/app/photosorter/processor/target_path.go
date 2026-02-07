package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/geocoding"
	"photo-sorter/internal/pkg/metadata"
)

func GetTargetPath(path string, exif *metadata.ExifData, cfg *config.Config) (string, *geocoding.CountryCity, error) {
	return GetTargetPathWithGeoResolver(path, exif, cfg, metadata.NewDefaultGeoResolver(cfg))
}

func GetTargetPathWithGeoResolver(path string, exif *metadata.ExifData, cfg *config.Config, geoResolver metadata.GeoResolver) (string, *geocoding.CountryCity, error) {
	if geoResolver == nil {
		geoResolver = metadata.NewDefaultGeoResolver(cfg)
	}
	return GetTargetPathWithResolvers(
		path,
		exif,
		cfg,
		geoResolver,
	)
}

func GetTargetPathWithResolvers(
	path string,
	exif *metadata.ExifData,
	cfg *config.Config,
	geoResolver metadata.GeoResolver,
) (string, *geocoding.CountryCity, error) {
	// 解析日期
	date, err := metadata.ResolveDate(exif, cfg)
	if err != nil {
		return "", nil, err
	}

	// 解析地理位置
	date, location, err := geoResolver.Resolve(date, exif, cfg)
	if err != nil {
		return "", nil, err
	}

	device := metadata.ResolveDevice(exif)

	// 建立目標路徑
	targetDir := filepath.Join(cfg.DstDir, date, device)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create target directory: %v", err)
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

	return targetPath, location, nil
}
