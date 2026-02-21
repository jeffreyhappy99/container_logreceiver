package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 用於自動加前綴 + timestamp
type prefixWriter struct {
	w      *os.File
	prefix string
	mu     sync.Mutex
}

func (p *prefixWriter) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	ts := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s %s%s", ts, p.prefix, string(data))
	n, err := p.w.Write([]byte(line))
	if err != nil {
		return n, err
	}
	return n, p.w.Sync() // flush 立即寫入
}

func main() {
	// 1. FD 對應 NewBinaryIO
	pipeStdout := os.NewFile(3, "pipeStdout")
	pipeStderr := os.NewFile(4, "pipeStderr")
	pipeSync := os.NewFile(5, "pipeSync")

	// 2. logPath
	logPath := filepath.Join(os.TempDir(), "container.log")
	if len(os.Args) > 1 {
		logPath = filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", os.Args[1]))
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "無法開啟日誌檔案: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// 3. handshake
	pipeSync.Close() // 通知 NewBinaryIO binary 已準備好

	// 4. 並發讀取 stdout/stderr
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		writer := &prefixWriter{w: f, prefix: "[STDOUT] "}
		scanner := bufio.NewScanner(pipeStdout)
		for scanner.Scan() {
			writer.Write([]byte(scanner.Text() + "\n"))
		}
	}()

	go func() {
		defer wg.Done()
		writer := &prefixWriter{w: f, prefix: "[STDERR] "}
		scanner := bufio.NewScanner(pipeStderr)
		for scanner.Scan() {
			writer.Write([]byte(scanner.Text() + "\n"))
		}
	}()

	wg.Wait() // 等到 container exit
}
