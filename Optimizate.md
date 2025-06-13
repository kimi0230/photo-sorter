主要的記憶體操作來自以下幾個地方：

1. `internal/app/photo-sorter/app.go` 中的 `copyFile` 函數：
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
