package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// prefixWriter 是一個包裝 os.File 的 writer，
// 會在每行前面加上時間戳與指定前綴，並確保寫入時 thread-safe。
type prefixWriter struct {
	w      *os.File   // 實際寫入的檔案
	prefix string     // 每行前綴（如 [STDOUT] 或 [STDERR]）
	mu     sync.Mutex // 保證多 goroutine 寫入時不會交錯
}

// Write 會加上時間戳與前綴後寫入檔案，並呼叫 Sync 立即 flush 到磁碟。
func (p *prefixWriter) Write(data []byte) (int, error) {

	defer p.mu.Unlock()
	// 取得目前時間字串
	ts := time.Now().Format("2006-01-02 15:04:05")
	// 組合完整 log 行：時間 前綴 內容
	line := fmt.Sprintf("%s %s%s", ts, p.prefix, string(data))
	// 寫入檔案
	n, err := p.w.Write([]byte(line))
	if err != nil {
		return n, err
	}
	// 立即 flush 到磁碟，確保 log 不會遺失
	return n, p.w.Sync()
}

func main() {
	// 取得由 containerd shim 傳來的三個 pipe：
	// fd 3: stdout, fd 4: stderr, fd 5: sync（啟動同步用）
	pipeStdout := os.NewFile(3, "pipeStdout")
	pipeStderr := os.NewFile(4, "pipeStderr")
	pipeSync := os.NewFile(5, "pipeSync")

	// 預設 log 檔案路徑為 /tmp/container.log，可由參數指定不同檔名
	logPath := filepath.Join(os.TempDir(), "container.log")
	if len(os.Args) > 1 {
		logPath = filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", os.Args[1]))
	}

	// 開啟 log 檔案（追加模式），失敗則輸出錯誤並結束
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "無法開啟日誌檔案: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// 關閉 sync pipe，通知 shim logging binary 已 ready
	pipeSync.Close()

	var wg sync.WaitGroup
	wg.Add(2) // 兩個 goroutine：分別處理 stdout 與 stderr

	// 處理 stdout：每行加上 [STDOUT] 前綴與時間戳後寫入 log 檔
	go func() {
		defer wg.Done()
		writer := &prefixWriter{w: f, prefix: "[STDOUT] "}
		scanner := bufio.NewScanner(pipeStdout)
		for scanner.Scan() {
			writer.Write([]byte(scanner.Text() + "\n"))
		}
	}()

	// 處理 stderr：每行加上 [STDERR] 前綴與時間戳後寫入 log 檔
	go func() {
		defer wg.Done()
		writer := &prefixWriter{w: f, prefix: "[STDERR] "}
		scanner := bufio.NewScanner(pipeStderr)
		for scanner.Scan() {
			writer.Write([]byte(scanner.Text() + "\n"))
		}
	}()

	// 等待兩個 goroutine 結束（即兩個 pipe 都關閉）
	wg.Wait()
}
