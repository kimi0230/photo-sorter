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
	LogLevel     string                 `yaml:"log_level"`      // 日誌等級：debug, info, warn, error
	EnableVerify bool                   `yaml:"enable_verify"`  // 是否啟用驗證
}

func LoadConfig(configPath string) (*Config, error) {
	// 讀取設定檔
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("讀取設定檔失敗: %v", err)
	}

	// 解析 YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析設定檔失敗: %v", err)
	}

	// 設定預設值
	if cfg.Workers == 0 {
		cfg.Workers = 4
	}
	if cfg.DstDir == "" {
		cfg.DstDir = cfg.SrcDir + "_sort"
	}
	if cfg.DryRun {
		cfg.DryRun = true
	}
	if cfg.DateFormat == "" {
		cfg.DateFormat = "2006-01"
	}
	if cfg.EnableGeoTag {
		cfg.EnableGeoTag = true
	}
	if cfg.GeoJSONPath == "" {
		cfg.GeoJSONPath = "./internal/pkg/geocoding/geodata/states.geojson"
	}
	if cfg.GeocoderType == "" {
		cfg.GeocoderType = geocoding.GeoStateType
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info" // 預設日誌等級為 info
	}

	return &cfg, nil
}

func (c *Config) ApplyFlags(srcDir, dstDir string, workers int) {
	// 如果命令列有指定參數，則覆蓋設定檔的值
	if srcDir != "." {
		c.SrcDir = srcDir
	}
	if dstDir == "." {
		baseName := filepath.Base(srcDir)
		c.DstDir = filepath.Join(".", baseName+"_sort")
	} else {
		c.DstDir = dstDir
	}

	if workers > 0 {
		c.Workers = workers
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
