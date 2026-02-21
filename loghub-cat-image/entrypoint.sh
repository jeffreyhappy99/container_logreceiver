#!/bin/sh
set -eu

# 預設讀這個檔案；你可以在 ctr run 時用 --env LOG_FILE=... 覆蓋
LOG_FILE="${LOG_FILE:-/logs/app.log}"

# 印一行提示到 stderr（方便 debug）
echo "[loghub-cat] reading: $LOG_FILE" >&2

# 檔案不存在就直接報錯
if [ ! -f "$LOG_FILE" ]; then
  echo "[loghub-cat] ERROR: file not found: $LOG_FILE" >&2
  exit 1
fi

# 先把目前所有內容印出，再持續追蹤新增內容
cat "$LOG_FILE"

#立刻把「那一行」寫到 stdout
exec tail -n 0 -F "$LOG_FILE"
