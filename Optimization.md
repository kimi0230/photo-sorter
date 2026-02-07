主要的記憶體操作來自以下幾個地方：

1. `internal/app/photosorter/app.go` 中的 `copyFile` 函數：
```go
func (a *App) copyFile(src, dst string) error {
    input, err := os.ReadFile(src)  // 這裡會讀取整個檔案到記憶體
    if err != nil {
        return err
    }

    err = os.WriteFile(dst, input, 0644)  // 這裡會寫入整個檔案
    if err != nil {
        return err
    }

    return nil
}
```

1. `internal/pkg/geocoding/geo_state.go` 中的 `loadGeoJSON` 函數：
```go
func (g *GeoState) loadGeoJSON() error {
    jsonFile, err := os.Open(g.jsonPath)
    if err != nil {
        return err
    }
    defer jsonFile.Close()

    byteValue, err := io.ReadAll(jsonFile)  // 這裡會讀取整個 JSON 檔案到記憶體
    if err != nil {
        return err
    }

    g.collection = &GeoJSONCollection{}
    if err := json.Unmarshal(byteValue, g.collection); err != nil {
        return err
    }

    return nil
}
```

建議的優化方案：

1. 對於檔案複製，使用緩衝區串流：
```go
func (a *App) copyFile(src, dst string) error {
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

    buffer := make([]byte, 32*1024) // 32KB 緩衝區
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
```

2. 對於 GeoJSON 資料，可以：
   - 使用記憶體映射（mmap）
   - 實作資料分頁載入
   - 使用資料庫儲存（如 SQLite）


---
其他方案: 替換成 **SpatiaLite**

---
## 2026-02-07

### 1. EXIF 架構重構與依賴注入
* **acd874d**：`ProcessFile` 改為接受 `exifReader`、`taggerProvider` 參數；.gitignore 加入暫存檔與 profiler。
* **80be0f9**：引入 `ExifReader` interface 與 legacy 實作；更新單元測試與 CI 加入 unit testing。
* **ab03dc5**：將 `ExiftoolClient` 整合進 `ProcessFile` 與 Worker，統一 EXIF 讀取流程。

### 2. EXIF 解析器（resolvers）
* **f453bfd**：為 EXIF 處理引入 resolvers，強化日期、裝置、地理資訊的提取（date / device / geo）。

### 3. 可自訂 EXIF reader 與 tagger 工廠（ce02e12）
* 支援可注入的 EXIF reader、tagger provider 工廠，提升測試性與替換彈性。

### 4. 錯誤處理與 API 簡化
* **a456ed4**：context 取消時的錯誤處理改為較完整的 error wrapping。
* **a80f597**：簡化 `HandleUnsupportedFile`、`HandelFailedFolder`，移除 logger 參數；更新 `ProcessFile`、Worker 的錯誤處理。

### 5. 版本與 job/result 流程（72733f0）
* 版本更新至 v0.1.23；重構 job 掃描與結果收集邏輯。

### 6. 測試與 CI
* **06e5668**：geocoder 測試加入 geodata 路徑解析、spatialite 指令檢查等輔助函式。
* **547f60f**：geocoder 與 mac tagger 測試的錯誤訊息標準化。
* **15bee0a**：CI 加入 macOS 安裝 `tag` 的步驟；單元測試在無 `tag` 時可略過。
* **326f93a**：Makefile 新增 `clean-test-data`；photo sorter 加入目錄驗證邏輯，提升處理正確性。

### 7. 其他
* **1b87385**：README 更新、中文在地化；日誌與錯誤訊息更清晰。
* **88a41a9**：DJI 機型對應加入 Osmo Pocket 3；ExifData 新增 Encoder 欄位。
* **ba366a2**：VSCode launch 設定改為 debug 模式並指定 main.go。

---
## 2026-02-08

### 1. 可自訂的 warning 記錄與忽略檔案日誌（088c2d7）
* **對應項目**：避免在 library 內直接 `fmt.Printf`；改成 logger 或回傳可被記錄的 warning，讓 UI/CLI 端控制輸出。
* **實作**：
  * metadata 處理的 warning 改為可注入的 `WarningLogger`，由呼叫端（如 `main.go`）決定如何輸出。
  * Worker 啟動時會記錄被忽略的檔案，便於除錯與統計。

### 2. Geo 快取與解析增強（3b85bed）
* **對應項目**：Geocoder 建立成本可下沉/快取；同座標或同批次重複查詢可避免重複建立與查詢。
* **實作**：
  * 新增 `GeoResolver` 與 geo 快取功能，可在 config 中啟用（如 `configs/config.yaml`）。
  * App 啟動時建立 resolver 並注入，避免每筆檔案都 `NewGeocoder`。
  * 座標解析結果可快取，減少重複查詢與 I/O。

### 3. 結構化結果類型（cd3986a）
* **對應項目**：結果通道只傳「錯誤」或改成結構化結果；後續統計/記錄更清楚。
* **實作**：
  * 改為傳遞結構化 result 類型（如 `struct{ path string; err error }`），不再只傳 `results <- err` 塞很多 nil。
  * `app.go` 與 `worker/worker.go` 依新類型處理，錯誤回報與統計更明確。

### 4. GetExifData 支援 context 取消（35f0b15）
* **對應項目**：GetExifData / ExiftoolClient 改用 `exec.CommandContext`，worker 的 ctx 取消時能中止 exiftool。
* **實作**：
  * `GetExifData` 改為接受 `context.Context`，內部使用 `exec.CommandContext` 執行 exiftool。
  * Worker 取消時可中止 exiftool 行程，避免卡住或長時間等待；錯誤傳遞一併整理。
  * `file/file.go`、`exif.go`、`exif_mock.go` 已配合更新。
