# 使用 exiftool 來分類照片與影片

## 分類方式
如果有地理位置
-> 根據時間-地理位置->裝置
```
sorted_media/
├── 2024-08-02-Japan/
│   └── GoPro_HERO8_Black/
│       └── GH011629.MP4
├── 2024-06-01/
│   └── iPhone11/
...
```

如果沒有地理位置
-> 根據時間->裝置

```
sorted_media/
├── 2024-08-02/
│   └── GoPro_HERO8_Black/
│       └── GH011629.MP4
├── 2024-06-01/
│   └── iPhone11/
...
```

* 依拍攝日期 (Create Date) 建立 YYYY-MM-DD 資料夾
* 依裝置名稱 (Camera Model Name) 建立子目錄，需將空格轉為底線，去除特殊字元
* 同一日期有多台裝置，落在同一日期資料夾中
* 支援的副檔名：照片（JPG、JPEG、HEIC、PNG）、影片（MP4、MOV）
* 若無 Create Date，可使用 Media Create Date 或 fallback 為 unknown_date
* 裝置名稱缺失時 fallback 為 unknown_device
* 同名檔案應避免覆蓋，使用自動加尾碼的方式避免衝突（如 _1, _2）
* 如無法處理，記錄錯誤與跳過，寫入 error.log
* 需支援更多維度分類（如 GPS、影片長度、拍攝模式等），可於日後擴充。

## 使用 Golang 實現
### 參數
如果有缺參數也幫我補上
| 參數   | 說明                             |
|--------|----------------------------------|
| `-src` | 原始照片資料夾，預設為當下目錄   |
| `-dst` | 整理後儲存的位置，預設為當下目錄 |

### 併發
* 使用 worker 模式，可設定最大併發數
* 有 logger 紀錄開始時間、每一筆處理結果、錯誤訊息
* 統計：起始檔案總數、成功分類數量、失敗數量
* 支援 graceful shutdown（SIGINT 時完成當前任務後結束）


（Optional）
* dry-run 模式：僅顯示將搬移的路徑，不實際執行

## 部署
* 提供 Makefile
* 提供 Dockerfile


每次的修改須連同 `README.md` 一起修改
