# 地理編碼套件 (Geocoding Package)

這個套件提供地理編碼功能，可以根據 GPS 座標（緯度和經度）查找對應的國家和城市信息。

## 功能特點

- 支援 GeoJSON 格式的地理數據
- 提供國家和城市級別的地理編碼
- 支援多邊形和多重多邊形的幾何形狀
- 使用射線法進行點在多邊形內的判斷

## 使用方式

```go
import "your-project/internal/pkg/geocoding"

// 創建地理編碼器實例
geocoder, err := geocoding.NewGeocoder(geocoding.GeoStateType, map[string]interface{}{
    "json_path": "path/to/your/geojson/file",
})
if err != nil {
    // 處理錯誤
}

// 根據 GPS 座標查找位置
location, err := geocoder.GetLocationFromGPS(25.0330, 121.5654)
if err != nil {
    // 處理錯誤
}

// 獲取國家和城市信息
country := location.Country  // 例如：TWN
city := location.City        // 例如：Taipei
```

## 數據來源

本套件使用以下數據源：

- [Natural Earth Data](https://www.naturalearthdata.com/) - 提供高質量的地理數據
- [Kaggle Country Coordinates Dataset](https://www.kaggle.com/datasets/danielvalyano/country-coord) - 提供國家座標數據

## 開發工具

### 地理數據處理

```sh
# 查看 shapefile 信息
ogrinfo ./vsizip/ne_50m_admin_1_states_provinces.shp -al -so
ogrinfo ./vsizip/ne_10m_admin_1_states_provinces.shp -al -so
```

### 性能分析

```sh
# 安裝 graphviz（用於生成性能分析圖表）
brew install graphviz

go run cmd/photo-sorter/main.go -c configs/config.yaml -cpuprofile cpu.prof

# 啟動性能分析 Web 界面
go tool pprof -http=:8080 cpu.prof
```

## 相關資源

- [GeoJSON 規範](https://geojson.org/)
- [Natural Earth Data 文檔](https://www.naturalearthdata.com/downloads/)
- [GDAL/OGR 文檔](https://gdal.org/programs/ogrinfo.html)
- https://www.kaggle.com/datasets/danielvalyano/country-coord?
resource=download
- https://github.co m/datasets/geo-countries/blob/main/Makefile
