# 地理編碼查詢工具

這是一個基於 Spatialite 的地理編碼查詢工具，可以查詢地理位置信息和數據庫表結構。

## 功能特色

- **位置查詢**：根據經緯度查詢行政區信息
- **表結構查詢**：查看 Spatialite 數據庫的表結構和統計信息
- **命令行界面**：支持靈活的命令行參數
- **多語言支持**：支持多種語言的表結構描述

## 安裝要求

- Go 1.12 或更高版本
- Spatialite 5.1.0 或更高版本
- Natural Earth 數據庫文件

## 檔案結構

```
cmd/geo/
├── main.go      # 主程序入口，處理命令行參數和路由
├── location.go  # 位置查詢相關功能
├── geometry.go  # 表結構查詢相關功能
└── README.md    # 本文件
```

## 使用方法

### 基本語法

```bash
go run ./cmd/geo [選項]
```

### 命令行參數

| 參數          | 類型    | 默認值                            | 說明                  |
|---------------|---------|-----------------------------------|-----------------------|
| `-db_path`    | string  | `states.sqlite`                   | Spatialite 數據庫路徑 |
| `-lat`        | float64 | `43.06417`                        | 緯度                  |
| `-lon`        | float64 | `141.34694`                       | 經度                  |
| `-cmd`        | string  | `location`                        | 執行的命令            |
| `-table_name` | string  | `ne_10m_admin_1_states_provinces` | 要查詢的表名          |
| `-help`       | bool    | `false`                           | 顯示幫助信息          |

### 命令類型

- `location`：查詢地理位置（默認）
- `geometry`：查詢表結構

## 使用範例

### 1. 查詢位置信息

```bash
# 查詢台北位置（預設座標）
go run ./cmd/geo

# 查詢指定座標
go run ./cmd/geo -lat=25.0330 -lon=121.5654

# 查詢東京
go run ./cmd/geo -lat=35.6895 -lon=139.6917

# 查詢紐約
go run ./cmd/geo -lat=40.7128 -lon=-74.0060
```

### 2. 查詢表結構

```bash
# 查詢默認表結構
go run ./cmd/geo -cmd=geometry

# 查詢指定表結構
go run ./cmd/geo -cmd=geometry -table_name=my_table

# 使用自定義數據庫查詢表結構
go run ./cmd/geo -cmd=geometry -db_path=my_database.sqlite -table_name=spatial_table
```

### 3. 顯示幫助

```bash
go run ./cmd/geo -help
```

## 輸出範例

### 位置查詢輸出

```
查詢位置: 緯度=25.0330, 經度=121.5654
使用數據庫: states.sqlite
---
查詢結果:
Taipei|Taipei|TWN
```

### 表結構查詢輸出

```
查詢表結構
使用數據庫: states.sqlite
---
表結構:
表信息:
├─ 表名: ne_10m_admin_1_states_provinces
├─ 類型: 空間表 (Spatial Table)
├─ 記錄數量: 4596 筆
├─ 描述: 自然地球 1:10M 行政區劃分 (州/省級別)
└─ 主要欄位:
   ├─ ogc_fid (INTEGER): 主鍵識別碼
   ├─ featurecla (VARCHAR): 文字資料
   ├─ scalerank (INTEGER): 整數資料
   ├─ name (VARCHAR): 名稱
   ├─ admin (VARCHAR): 行政區名稱
   ├─ adm0_a3 (VARCHAR): 國家代碼 (ISO 3166-1 alpha-3)
   ├─ latitude (FLOAT): 緯度
   ├─ longitude (FLOAT): 經度
   └─ GEOMETRY (MULTIPOLYGON): 空間幾何數據
```

## 數據庫表結構

### ne_10m_admin_1_states_provinces 表

這是 Natural Earth 1:10M 行政區劃分表，包含以下主要欄位：

| 欄位名     | 類型         | 說明                          |
|------------|--------------|-------------------------------|
| ogc_fid    | INTEGER      | 主鍵識別碼                    |
| featurecla | VARCHAR(24)  | 特徵分類                      |
| scalerank  | INTEGER      | 縮放等級                      |
| adm1_code  | VARCHAR(9)   | 行政區代碼                    |
| name       | VARCHAR(44)  | 地區名稱                      |
| admin      | VARCHAR(36)  | 行政區名稱                    |
| adm0_a3    | VARCHAR(3)   | 國家代碼 (ISO 3166-1 alpha-3) |
| latitude   | FLOAT        | 緯度                          |
| longitude  | FLOAT        | 經度                          |
| GEOMETRY   | MULTIPOLYGON | 空間幾何數據                  |

完整的表結構（共 123 個欄位）：

```sql
PRAGMA table_info(ne_10m_admin_1_states_provinces);

0|ogc_fid|INTEGER|0||1
1|featurecla|VARCHAR(24)|0||0
2|scalerank|INTEGER|0||0
3|adm1_code|VARCHAR(9)|0||0
4|diss_me|INTEGER|0||0
5|iso_3166_2|VARCHAR(8)|0||0
6|wikipedia|VARCHAR(84)|0||0
7|iso_a2|VARCHAR(2)|0||0
8|adm0_sr|INTEGER|0||0
9|name|VARCHAR(44)|0||0
10|name_alt|VARCHAR(129)|0||0
11|name_local|VARCHAR(66)|0||0
12|type|VARCHAR(38)|0||0
13|type_en|VARCHAR(27)|0||0
14|code_local|VARCHAR(5)|0||0
15|code_hasc|VARCHAR(8)|0||0
16|note|VARCHAR(114)|0||0
17|hasc_maybe|VARCHAR(13)|0||0
18|region|VARCHAR(43)|0||0
19|region_cod|VARCHAR(15)|0||0
20|provnum_ne|INTEGER|0||0
21|gadm_level|INTEGER|0||0
22|check_me|INTEGER|0||0
23|datarank|INTEGER|0||0
24|abbrev|VARCHAR(9)|0||0
25|postal|VARCHAR(3)|0||0
26|area_sqkm|INTEGER|0||0
27|sameascity|INTEGER|0||0
28|labelrank|INTEGER|0||0
29|name_len|INTEGER|0||0
30|mapcolor9|INTEGER|0||0
31|mapcolor13|INTEGER|0||0
32|fips|VARCHAR(5)|0||0
33|fips_alt|VARCHAR(9)|0||0
34|woe_id|INTEGER|0||0
35|woe_label|VARCHAR(64)|0||0
36|woe_name|VARCHAR(44)|0||0
37|latitude|FLOAT|0||0
38|longitude|FLOAT|0||0
39|sov_a3|VARCHAR(3)|0||0
40|adm0_a3|VARCHAR(3)|0||0
41|adm0_label|INTEGER|0||0
42|admin|VARCHAR(36)|0||0
43|geonunit|VARCHAR(40)|0||0
44|gu_a3|VARCHAR(3)|0||0
45|gn_id|INTEGER|0||0
46|gn_name|VARCHAR(72)|0||0
47|gns_id|INTEGER|0||0
48|gns_name|VARCHAR(80)|0||0
49|gn_level|INTEGER|0||0
50|gn_region|VARCHAR(1)|0||0
51|gn_a1_code|VARCHAR(10)|0||0
52|region_sub|VARCHAR(41)|0||0
53|sub_code|VARCHAR(5)|0||0
54|gns_level|INTEGER|0||0
55|gns_lang|VARCHAR(3)|0||0
56|gns_adm1|VARCHAR(4)|0||0
57|gns_region|VARCHAR(4)|0||0
58|min_label|FLOAT|0||0
59|max_label|FLOAT|0||0
60|min_zoom|FLOAT|0||0
61|wikidataid|VARCHAR(9)|0||0
62|name_ar|VARCHAR(85)|0||0
63|name_bn|VARCHAR(134)|0||0
64|name_de|VARCHAR(50)|0||0
65|name_en|VARCHAR(47)|0||0
66|name_es|VARCHAR(44)|0||0
67|name_fr|VARCHAR(47)|0||0
68|name_el|VARCHAR(85)|0||0
69|name_hi|VARCHAR(134)|0||0
70|name_hu|VARCHAR(47)|0||0
71|name_id|VARCHAR(46)|0||0
72|name_it|VARCHAR(47)|0||0
73|name_ja|VARCHAR(96)|0||0
74|name_ko|VARCHAR(54)|0||0
75|name_nl|VARCHAR(46)|0||0
76|name_pl|VARCHAR(45)|0||0
77|name_pt|VARCHAR(43)|0||0
78|name_ru|VARCHAR(85)|0||0
79|name_sv|VARCHAR(41)|0||0
80|name_tr|VARCHAR(44)|0||0
81|name_vi|VARCHAR(71)|0||0
82|name_zh|VARCHAR(61)|0||0
83|ne_id|BIGINT|0||0
84|name_he|VARCHAR(63)|0||0
85|name_uk|VARCHAR(89)|0||0
86|name_ur|VARCHAR(103)|0||0
87|name_fa|VARCHAR(92)|0||0
88|name_zht|VARCHAR(61)|0||0
89|fclass_iso|VARCHAR(12)|0||0
90|fclass_us|VARCHAR(12)|0||0
91|fclass_fr|VARCHAR(1)|0||0
92|fclass_ru|VARCHAR(12)|0||0
93|fclass_es|VARCHAR(12)|0||0
94|fclass_cn|VARCHAR(18)|0||0
95|fclass_tw|VARCHAR(12)|0||0
96|fclass_in|VARCHAR(12)|0||0
97|fclass_np|VARCHAR(12)|0||0
98|fclass_pk|VARCHAR(12)|0||0
99|fclass_de|VARCHAR(12)|0||0
100|fclass_gb|VARCHAR(12)|0||0
101|fclass_br|VARCHAR(12)|0||0
102|fclass_il|VARCHAR(12)|0||0
103|fclass_ps|VARCHAR(12)|0||0
104|fclass_sa|VARCHAR(12)|0||0
105|fclass_eg|VARCHAR(12)|0||0
106|fclass_ma|VARCHAR(1)|0||0
107|fclass_pt|VARCHAR(12)|0||0
108|fclass_ar|VARCHAR(12)|0||0
109|fclass_jp|VARCHAR(12)|0||0
110|fclass_ko|VARCHAR(12)|0||0
111|fclass_vn|VARCHAR(12)|0||0
112|fclass_tr|VARCHAR(1)|0||0
113|fclass_id|VARCHAR(12)|0||0
114|fclass_pl|VARCHAR(1)|0||0
115|fclass_gr|VARCHAR(12)|0||0
116|fclass_it|VARCHAR(12)|0||0
117|fclass_nl|VARCHAR(1)|0||0
118|fclass_se|VARCHAR(12)|0||0
119|fclass_bd|VARCHAR(12)|0||0
120|fclass_ua|VARCHAR(12)|0||0
121|fclass_tlc|VARCHAR(12)|0||0
122|GEOMETRY|MULTIPOLYGON|0||0
```

## 技術細節

### 空間查詢

工具使用 Spatialite 的空間函數進行地理位置查詢：

```sql
SELECT name, admin, adm0_a3 FROM ne_10m_admin_1_states_provinces
WHERE ST_Contains(
    GEOMETRY,
    ST_PointFromText('POINT(121.5654 25.0330)', 4326)
);
```

### 支持的坐標系統

- **EPSG:4326**：WGS84 地理坐標系統（經度/緯度）

### 錯誤處理

- 數據庫連接錯誤
- 空間查詢錯誤
- 表不存在錯誤
- 參數驗證錯誤

## 開發說明

### 添加新命令

1. 在 `main.go` 中添加新的命令常量
2. 創建新的處理函數
3. 在 switch 語句中添加新的 case

### 擴展功能

- 支持更多數據庫格式
- 添加更多查詢類型
- 支持批量查詢
- 添加結果導出功能

## 故障排除

### 常見問題

1. **找不到 spatialite 執行檔**
   - 確保 Spatialite 已正確安裝
   - 檢查 PATH 環境變數

2. **數據庫文件不存在**
   - 檢查 `-db_path` 參數指定的路徑
   - 確保文件有讀取權限

3. **查詢結果為空**
   - 檢查坐標是否在數據庫覆蓋範圍內
   - 確認數據庫包含相應的空間數據

### 調試模式

可以通過修改代碼添加更詳細的錯誤信息輸出。

## 授權

本工具遵循項目的整體授權條款。

## 貢獻

歡迎提交 Issue 和 Pull Request 來改進這個工具。
