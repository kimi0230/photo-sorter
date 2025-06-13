#!/bin/bash

# =============================================
# 檔案計數腳本
# =============================================
# 功能：計算指定目錄中的檔案數量
#
# 參數：
#   $1: 目錄路徑 - 要計算檔案數量的目錄
#
# 輸出：
#   - 總檔案數量
#   - 不包含 ._ 開頭檔案的數量
#
# 使用方式：
#   ./count_files.sh <目錄路徑>
#
# 範例：
#   ./count_files.sh /path/to/your/directory
# =============================================

# 檢查是否提供了目錄路徑
if [ -z "$1" ]; then
    echo "錯誤：請提供目錄路徑"
    echo "使用方式：./count_files.sh <目錄路徑>"
    exit 1
fi

# 檢查目錄是否存在
if [ ! -d "$1" ]; then
    echo "錯誤：目錄 '$1' 不存在"
    exit 1
fi

# 計算檔案數量
file_count=$(find "$1" -type f | wc -l)

# 輸出結果
echo "目錄 '$1' 中的檔案數量：$file_count" 

file_count_no_dot=$(find "$1" -type f -not -name "._*" | wc -l)
echo "目錄 '$1' 中的檔案數量（不包含 ._ 開頭的檔案）：$file_count_no_dot"
