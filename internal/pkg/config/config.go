package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/geocoding"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SrcDir       string                 `yaml:"src_dir"`
	DstDir       string                 `yaml:"dst_dir"`
	Workers      int                    `yaml:"workers"`
	DryRun       bool                   `yaml:"dry_run"`
	Ignore       []string               `yaml:"ignore"`         // 要忽略的檔案類型
	Formats      []string               `yaml:"formats"`        // 支援的檔案格式
	DateFormat   string                 `yaml:"date_format"`    // 日期格式：YYYY-MM-DD 或 YYYY-MM
	EnableGeoTag bool                   `yaml:"enable_geo_tag"` // 是否啟用地理位置標籤
	GeoJSONPath  string                 `yaml:"geo_json_path"`  // GeoJSON 檔案路徑
	GeocoderType geocoding.GeocoderType `yaml:"geocoder_type"`  // 地理編碼器類型
}

func LoadConfig() (*Config, error) {
	// 預設設定
	config := &Config{
		SrcDir:       ".",
		DstDir:       "sorted_media",
		Workers:      4,
		DryRun:       false,
		DateFormat:   "2006-01-02",                                  // 預設使用完整日期
		EnableGeoTag: true,                                          // 預設啟用地理位置標籤
		GeoJSONPath:  "./internal/pkg/geocoding/countries.geo.json", // 預設 GeoJSON 路徑
		GeocoderType: geocoding.GeoAlpha3JSONType,                   // 預設使用 GeoAlpha3JSON 地理編碼器
		Ignore: []string{
			".git", ".gitignore",
			".go", ".mod", ".sum",
			".md", ".log", ".yaml",
			".sample",
		},
		Formats: []string{
			".jpg", ".jpeg", ".heic", ".png",
			".mp4", ".mov",
		},
	}

	// 嘗試讀取設定檔
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("讀取設定檔失敗: %v", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("解析設定檔失敗: %v", err)
		}
	}

	return config, nil
}

func (c *Config) ApplyFlags(srcDir, dstDir string, workers int, dryRun bool) {
	// 如果命令列有指定參數，則覆蓋設定檔的值
	if srcDir != "." {
		c.SrcDir = srcDir
	}
	if dstDir != "." {
		c.DstDir = dstDir
	}
	if workers != 4 {
		c.Workers = workers
	}
	if dryRun {
		c.DryRun = true
	}
}

func (c *Config) ShouldIgnore(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	baseName := strings.ToLower(filepath.Base(path))

	// 檢查副檔名和檔案名是否在忽略清單中
	for _, ignore := range c.Ignore {
		ignore = strings.ToLower(ignore)
		if strings.HasPrefix(ignore, ".") {
			// 如果是副檔名，檢查檔案副檔名
			if ext == ignore {
				return true
			}
		} else {
			// 如果是檔案名，檢查完整檔案名
			if baseName == ignore {
				return true
			}
		}
	}
	return false
}

func (c *Config) IsSupportedFormat(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, format := range c.Formats {
		if ext == strings.ToLower(format) {
			return true
		}
	}
	return false
}

func CreateDefaultConfig() error {
	config := &Config{
		SrcDir:       ".",
		DstDir:       "sorted_media",
		Workers:      4,
		DryRun:       false,
		DateFormat:   "2006-01-02",
		EnableGeoTag: true,
		GeoJSONPath:  "./internal/pkg/geocoding/countries.geo.json",
		GeocoderType: geocoding.GeoAlpha3JSONType,
		Ignore: []string{
			".git", ".gitignore",
			".go", ".mod", ".sum",
			".md", ".log", ".yaml",
			".sample",
		},
		Formats: []string{
			".jpg", ".jpeg", ".heic", ".png",
			".mp4", ".mov",
		},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("產生預設設定檔失敗: %v", err)
	}

	return os.WriteFile("config.yaml", data, 0644)
}
