# Photo Sorter

這是一個使用 Go 語言開發的照片和影片分類工具，可以根據拍攝時間、裝置型號和地理位置自動整理媒體檔案。

**目前只支援 MacOS**

## 功能特點

- 根據拍攝日期（Create Date）自動分類
- 支援多種媒體格式（JPG、JPEG、HEIC、PNG、MP4、MOV、GIF、WEBP、HEIF、HEVC、MKV、AVI、WMV、FLV、MPEG、MPG、M4V、BMP，以及 RAW 格式：CR2、CR3、NEF、NRW、ARW、RAF、RW2、ORF、PEF、DNG、RWL、RAW）
- 自動處理檔案名稱衝突
- 支援多工處理
- 提供詳細的處理日誌
- 支援優雅關閉（Graceful Shutdown）
- 支援地理位置標記（Geo Tagging）
- 提供詳細的處理統計資訊
- 命令列介面，支援靈活的選項
- 效能分析支援（CPU 和記憶體）

## 系統需求

- Go 1.23 或更高版本
- [exiftool](#exiftool)
- [spatialite](#spatialite-tools)
- [tag](#tag)

## 安裝

### 使用 Makefile

```bash
make build
```

### 使用 Docker

```bash
docker build -t photo-sorter .
```

### spatialite-tools

```sh
brew install spatialite-tools

which spatialite
```

### exiftool
```sh
brew install exiftool
```

### tag
```sh
brew install tag
```

### gpmf-parser

https://github.com/gopro/gpmf-parser/tree/main

```sh
git submodule add https://github.com/gopro/gpmf-parser.git third_party/gpmf-parser
git submodule update --init --recursive

cd third_party/gpmf-parser
mkdir build
cd build
cmake ..
make
chmod +x gpmf-parser
```

## 使用方法

### 基本使用

**建議先關閉 Spotlight 在執行檔案!!!**

```sh
# 暫停 Spotlight：
sudo mdutil -i off /

# 執行（指定來源目錄）
./photo-sorter -src /path/to/source/folder

# 執行（指定來源和目標目錄）
./photo-sorter -src /path/to/source/folder -dst /path/to/destination/folder

# 執行（指定工作執行緒數量）
./photo-sorter -src /path/to/source/folder -workers 8

# 執行（使用自訂配置檔）
./photo-sorter -src /path/to/source/folder -c /path/to/config.yaml

# 顯示版本資訊
./photo-sorter -version

# 完成後，重建 Spotlight index：
sudo mdutil -E /
sudo mdutil -i on /

```

### 命令列參數

| 參數          | 類型   | 預設值             | 說明                                                    |
|---------------|--------|--------------------|---------------------------------------------------------|
| `-src`        | string | `.`                | 原始照片資料夾路徑                                      |
| `-dst`        | string | `.`                | 整理後儲存的位置（如果為 `.`，則使用 `{src_dir}_sort`） |
| `-workers`    | int    | `runtime.NumCPU()` | 最大併發數                                              |
| `-c`          | string | `config.yaml`      | 配置檔案路徑                                            |
| `-version`    | bool   | `false`            | 顯示版本資訊                                            |
| `-cpuprofile` | string | ``                 | CPU profile 檔案路徑（用於效能分析）                    |
| `-memprofile` | string | ``                 | 記憶體 profile 檔案路徑（用於效能分析）                 |

### 配置檔案說明

```yaml
version: "0.1.22"  # 版本號

# 照片分類工具設定檔

# 原始照片資料夾路徑
src_dir: "source_media"

# 整理後儲存的位置（如果為空，預設為 {src_dir}_sort）
dst_dir: ""

# 併發工作執行緒數量（預設：4）
workers: 4

# 是否為乾跑模式（只顯示將要移動的檔案，不實際執行）
dry_run: false

# 日期格式：YYYY-MM-DD (2006-01-02) 或 YYYY-MM (2006-01)
date_format: "2006-01"

# 是否啟用地理位置標籤
enable_geo_tag: true

# Geo 資料庫檔案路徑（支援 .geojson 或 .sqlite）
# 選項：./geodata/states.geojson 或 ./geodata/states.sqlite
geo_db_path: "./geodata/states.sqlite"

# 地理編碼器類型："geo_json" 或 "geo_spatialite"
geocoder_type: "geo_spatialite"

# 日誌等級設定 (debug, info, warn, error)
log_level: "info"

# 是否啟用驗證
enable_verify: true

# 支援的檔案格式
formats:
  - ".jpg"
  - ".jpeg"
  - ".heic"
  - ".png"
  - ".mp4"
  - ".mov"
  - ".gif"
  - ".webp"
  - ".heif"
  - ".hevc"
  - ".mkv"
  - ".avi"
  - ".wmv"
  - ".flv"
  - ".mpeg"
  - ".mpg"
  - ".m4v"
  - ".bmp"
  - ".cr2"
  - ".cr3"
  - ".nef"
  - ".nrw"
  - ".arw"
  - ".raf"
  - ".rw2"
  - ".orf"
  - ".pef"
  - ".dng"
  - ".rwl"
  - ".raw"
  
# 要忽略的檔案類型
ignore:
  - ".git"
  - ".gitignore"
  - ".go"
  - ".mod"
  - ".sum"
  - ".md"
  - ".log"
  - ".yaml"
  - ".sample"
  - ".DS_Store"
  - "Thumbs.db"

```

### 使用 Docker

```bash
docker run -v /path/to/photos:/app/input -v /path/to/output:/app/output photo-sorter -config config.yaml
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

- 日誌檔案會記錄在 logs/app.log
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
