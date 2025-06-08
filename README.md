# Photo Sorter

這是一個使用 Go 語言開發的照片和影片分類工具，可以根據拍攝時間和裝置型號自動整理媒體檔案。

## 功能特點

- 根據拍攝日期（Create Date）自動分類
- 支援多種媒體格式（JPG、JPEG、HEIC、PNG、MP4、MOV）
- 自動處理檔案名稱衝突
- 支援多工處理
- 提供詳細的處理日誌
- 支援優雅關閉（Graceful Shutdown）

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
./photo-sorter -src /path/to/photos -dst /path/to/output
```

### 參數說明

| 參數       | 說明                           | 預設值   |
|------------|--------------------------------|----------|
| `-src`     | 原始照片資料夾                 | 當前目錄 |
| `-dst`     | 整理後儲存的位置               | 當前目錄 |
| `-workers` | 最大併發數                     | 4        |
| `-dry-run` | 僅顯示將搬移的路徑，不實際執行 | false    |

### 使用 Docker

```bash
docker run -v /path/to/photos:/input -v /path/to/output:/output photo-sorter
```

## 輸出結構

```
sorted_media/
├── 2024-08-02/
│   └── GoPro_HERO8_Black/
│       └── GH011629.MP4
├── 2024-06-01/
│   └── iPhone11/
...
```

## 錯誤處理

- 無法處理的檔案會記錄在 error.log
- 缺少日期資訊的檔案會被歸類到 unknown_date 資料夾
- 缺少裝置資訊的檔案會被歸類到 unknown_device 資料夾

## 授權

MIT License
