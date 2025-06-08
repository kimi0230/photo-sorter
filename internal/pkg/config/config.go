package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SrcDir  string `yaml:"src_dir"`
	DstDir  string `yaml:"dst_dir"`
	Workers int    `yaml:"workers"`
	DryRun  bool   `yaml:"dry_run"`
}

func LoadConfig() (*Config, error) {
	// 預設設定
	config := &Config{
		SrcDir:  ".",
		DstDir:  ".",
		Workers: 4,
		DryRun:  false,
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

func CreateDefaultConfig() error {
	config := &Config{
		SrcDir:  ".",
		DstDir:  ".",
		Workers: 4,
		DryRun:  false,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("產生預設設定檔失敗: %v", err)
	}

	return os.WriteFile("config.yaml", data, 0644)
}
