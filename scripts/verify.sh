#!/bin/bash

# =============================================
# 目錄比對腳本
# =============================================
# 功能：比對兩個目錄中的檔案差異
#
# 參數：
#   $1: 來源目錄 - 要比對的第一個目錄
#   $2: 目標目錄 - 要比對的第二個目錄
#
# 輸出：
#   - 顯示只在來源目錄存在的檔案
#   - 顯示只在目標目錄存在的檔案
#
# 使用方式：
#   ./verify.sh <來源目錄> <目標目錄>
#
# 範例：
#   ./verify.sh /path/to/source /path/to/target
#
# 注意事項：
#   - 比對時會忽略檔案路徑，只比較檔名
#   - 會自動去除重複的檔名
#   - 使用 comm 命令進行比對
# =============================================

# 檢查參數數量
if [ "$#" -ne 2 ]; then
    echo "錯誤：需要兩個目錄路徑參數"
    echo "使用方式：./verify.sh <來源目錄> <目標目錄>"
    exit 1
fi

DIR_A="$1"
DIR_B="$2"

# 檢查目錄是否存在
if [ ! -d "$DIR_A" ]; then
    echo "錯誤：來源目錄 '$DIR_A' 不存在"
    exit 1
fi

if [ ! -d "$DIR_B" ]; then
    echo "錯誤：目標目錄 '$DIR_B' 不存在"
    exit 1
fi

./scripts/count_files.sh "$DIR_A"
./scripts/count_files.sh "$DIR_B"

echo "比對目錄："
echo "來源目錄：$DIR_A"
echo "目標目錄：$DIR_B"
echo "----------------------------------------"

# 列出 A 目錄下所有檔名（去掉路徑）
cd "$DIR_A" && find . -type f -exec basename {} \; | sort | uniq > /tmp/dirA_files.txt
cd "$DIR_B" && find . -type f -exec basename {} \; | sort | uniq > /tmp/dirB_files.txt

# 比對
echo "檔案差異："
comm -3 /tmp/dirA_files.txt /tmp/dirB_files.txt

# 清理臨時檔案
rm -f /tmp/dirA_files.txt /tmp/dirB_files.txt
