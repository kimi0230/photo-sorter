# Photo Sorter

這是一個使用 Go 語言開發的照片和影片分類工具，可以根據拍攝時間、裝置型號和地理位置自動整理媒體檔案。

## 功能特點

- 根據拍攝日期（Create Date）自動分類
- 支援多種媒體格式（JPG、JPEG、HEIC、PNG、MP4、MOV）
- 自動處理檔案名稱衝突
- 支援多工處理
- 提供詳細的處理日誌
- 支援優雅關閉（Graceful Shutdown）
- 支援地理位置標記（Geo Tagging）
- 提供詳細的處理統計資訊

## 系統需求

- Go 1.20 或更高版本
- exiftool

## 安裝

### 使用 Makefile

```bash
make build
```

### 使用 Docker

```bash
docker build -t photo-sorter .
```

## 使用方法

### 基本使用

```bash
./photo-sorter -config config/config.yaml
```

### 配置檔案說明

```yaml
# 基本設定
src_dir: "/path/to/photos"        # 原始照片資料夾
dst_dir: "/path/to/output"        # 整理後儲存的位置
max_workers: 4                    # 最大併發數
dry_run: false                    # 僅顯示將搬移的路徑，不實際執行
date_format: "2006-01-02"        # 日期格式

# 地理位置標記設定
enable_geo_tagging: true          # 啟用地理位置標記
geo_json_path: "./internal/pkg/geocoding/countries.geo.json"  # GeoJSON 檔案路徑
geocoder_type: "geo_alpha3_json"  # 地理編碼器類型

# 檔案格式設定
supported_formats:                # 支援的檔案格式
  - ".jpg"
  - ".jpeg"
  - ".heic"
  - ".png"
  - ".mp4"
  - ".mov"

ignore_files:                     # 忽略的檔案
  - ".DS_Store"
  - "Thumbs.db"
```

### 使用 Docker

```bash
docker run -v /path/to/photos:/app/input -v /path/to/output:/app/output photo-sorter -config config/config.yaml
```

## 輸出結構

```
sorted_media/
├── 2024-08-02-Japan/
│   └── GoPro_HERO8_Black/
│       └── GH011629.MP4
├── 2024-06-01/
│   └── iPhone11/
│       └── IMG_1234.JPG
├── unknown_date/
│   └── unknown_device/
│       └── IMG_5678.JPG
└── unknown_format/
    └── document.pdf
```

## 錯誤處理

- 無法處理的檔案會記錄在 error.log
- 缺少日期資訊的檔案會被歸類到 unknown_date 資料夾
- 缺少裝置資訊的檔案會被歸類到 unknown_device 資料夾
- 不支援的檔案格式會被歸類到 unknown_format 資料夾

## 處理統計

程式會提供詳細的處理統計資訊：
- 總檔案數
- 成功處理的檔案數
- 處理失敗的檔案數
- 處理時間
- 不支援的檔案格式統計
- 目錄結構及檔案數量統計
