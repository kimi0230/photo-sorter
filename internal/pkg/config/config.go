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
	SrcDir            string                 `yaml:"src_dir"`
	DstDir            string                 `yaml:"dst_dir"`
	Workers           int                    `yaml:"workers"`
	DryRun            bool                   `yaml:"dry_run"`
	Ignore            []string               `yaml:"ignore"`              // 要忽略的檔案類型
	Formats           []string               `yaml:"formats"`             // 支援的檔案格式
	DateFormat        string                 `yaml:"date_format"`         // 日期格式：YYYY-MM-DD 或 YYYY-MM
	EnableGeoTag      bool                   `yaml:"enable_geo_tag"`      // 是否啟用地理位置標籤
	EnableGeoCache    bool                   `yaml:"enable_geo_cache"`    // 是否啟用地理位置快取
	GeoCachePrecision int                    `yaml:"geo_cache_precision"` // 快取經緯度小數位
	GeoDBPath         string                 `yaml:"geo_db_path"`         // Geo DB 檔案路徑
	GeocoderType      geocoding.GeocoderType `yaml:"geocoder_type"`       // 地理編碼器類型
	LogLevel          string                 `yaml:"log_level"`           // 日誌等級：debug, info, warn, error
	EnableVerify      bool                   `yaml:"enable_verify"`       // 是否啟用驗證
}

func DefaultConfig() Config {
	return Config{
		Workers:           4,
		DateFormat:        "2006-01",
		GeoDBPath:         "./geodata/states.sqlite",
		GeocoderType:      geocoding.GeoTypeSpatialite,
		GeoCachePrecision: 2,
		LogLevel:          "info",
	}
}

func (c *Config) ApplyDefaults() {
	defaults := DefaultConfig()
	if c.Workers <= 0 {
		c.Workers = defaults.Workers
	}
	if c.DstDir == "" && c.SrcDir != "" {
		c.DstDir = c.SrcDir + "_sort"
	}
	if c.DateFormat == "" {
		c.DateFormat = defaults.DateFormat
	}
	if c.GeoDBPath == "" {
		c.GeoDBPath = defaults.GeoDBPath
	}
	if c.GeocoderType == "" {
		c.GeocoderType = defaults.GeocoderType
	}
	if c.GeoCachePrecision <= 0 {
		c.GeoCachePrecision = defaults.GeoCachePrecision
	}
	if c.LogLevel == "" {
		c.LogLevel = defaults.LogLevel
	}
}

func LoadConfig(configPath string) (*Config, error) {
	// 讀取設定檔
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析 YAML
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	cfg.ApplyDefaults()

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
