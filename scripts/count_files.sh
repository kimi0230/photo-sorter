#!/bin/bash

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
