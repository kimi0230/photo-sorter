

# TODO 優化

---

## ✅ 優化重點方向

### 1. **避免每次都反序列化 coordinates**

目前每個 feature 都會做：

```go
json.Unmarshal(feature.Geometry.Coordinates, ...)
```

這個步驟在大量資料下效能會變差，建議在 `loadGeoJSON()` 時就先解析完，改成：

#### ✅ 建議結構重構

```go
type GeometryPolygon struct {
	Type        string          `json:"type"`
	Coordinates interface{}     `json:"coordinates"`
}
```

然後在 `loadGeoJSON` 裡依據 `Type` 預先轉換成 `[][][]float64` 或 `[][][][]float64`，避免在每次查詢時都反序列化。

---

### 2. **加上 spatial index / bounding box 預過濾（可選，提升查詢效率）**

每筆 polygon 都可加上 Bounding Box（Envelope），初步過濾掉落在 bbox 外的查詢點，減少不必要的幾何運算。

#### 加欄位（在 `GeoJSONFeature`）：

```go
BBox [4]float64 // [minLon, minLat, maxLon, maxLat]
```

並在 `loadGeoJSON` 時解析：

```go
func calculateBBox(coords [][]float64) [4]float64 {
	// 算出最小/最大經緯度
}
```

然後在 `GetLocationFromGPS()` 內先做：

```go
if !pointInBBox(lat, lon, feature.BBox) {
    continue
}
```

---

### 3. **改成用 goroutine + channel 查詢（可選，加速多邊形判斷）**

如果 GeoJSON 有幾千筆 polygon，可考慮使用 goroutine + channel 並行判斷是否包含點。

這需保證：

* 查詢為讀取操作（不可同時寫入）
* 回傳第 1 個符合的即可取消其他 goroutine

---

### 4. **增加 `NameEn`、`ISOCode` 等欄位彈性**

你目前只有 `Name` 與 `Adm0A3`，如果你未來要用多語系或輸出格式，建議也能解析：

* `name_en` / `name_zh` 等語系欄位
* ISO Alpha-2 (`iso_a2`)
* `admin` 欄位作為國名

---

### 5. **測試與診斷強化**

增加以下功能能協助 debug 與效能分析：

```go
func (g *GeoState) Stats() string {
	return fmt.Sprintf("Loaded %d features from %s", len(g.collection.Features), g.jsonPath)
}
```

---

## ✨ Bonus：建立記憶體快取（如查詢重複點）

若你會重複查詢同一個座標組合，可加上簡單的 cache：

```go
type GeoState struct {
	cache map[string]*CountryCity
}

func (g *GeoState) GetLocationFromGPS(lat, lon float64) (*CountryCity, error) {
	key := fmt.Sprintf("%.6f,%.6f", lat, lon)
	if cc, ok := g.cache[key]; ok {
		return cc, nil
	}
	// 正常邏輯...
}
```

---

## 🧠 結論整理：

| 優化方向              | 建議                      |
|-----------------------|---------------------------|
| JSON 結構預解析       | ✅ 在載入時 parse geometry |
| Bounding Box 快速過濾 | ✅ 加速查詢效能            |
| 多線程查詢（可選）    | ⚠️ 適用於高併發/大資料集   |
| 支援多語系欄位        | ✅ 若輸出顯示有需求        |
| 建立座標快取（可選）  | ✅ 若查詢點經常重複        |

---
# 替換成 **SpatiaLite**
