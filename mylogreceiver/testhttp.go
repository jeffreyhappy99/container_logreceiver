package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// ---------------------- 新增：解析 log-uri query url ------------------------

var defaultURL = "http://128.9.0.1:8080/log"

func resolveTargetURL() string {

	for i := 0; i < len(os.Args)-1; i++ {
        if os.Args[i] == "url" {
            return os.Args[i+1]
        }
    }

	// fallback 預設值
	return defaultURL
}

// ---------------------------------------------------------------------------

// writer：只負責直接寫入內容，不做任何加工
type plainWriter struct {
	w  *os.File
	mu sync.Mutex
}

func (p *plainWriter) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n, err := p.w.Write(data)
	if err != nil {
		return n, err
	}
	return n, p.w.Sync()
}

// 把一行 log 用 HTTP POST 送到指定 URL
func sendLog(target string, line string) {
	resp, err := http.Post(target, "text/plain", bytes.NewBufferString(line+"\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sendLog error: %v\n", err)
		return
	}
	resp.Body.Close()
}

func main() {

	// 避免 flag 亂吃 os.Args
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// 重要：在一開始就解析 URL
	targetURL := resolveTargetURL()
    debugFile, _ := os.OpenFile("/tmp/testhttp-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    fmt.Fprintf(debugFile, ">>>> targetURL = %s\n", targetURL)
    fmt.Fprintf(debugFile, ">>>> os.Args = %#v\n", os.Args)
    debugFile.Sync()

	pipeStdout := os.NewFile(3, "pipeStdout")
	pipeStderr := os.NewFile(4, "pipeStderr")
	pipeSync := os.NewFile(5, "pipeSync")

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

	pipeSync.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	writer := &plainWriter{w: f}

	// stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(pipeStdout)
		for scanner.Scan() {
			line := scanner.Text()
			writer.Write([]byte(line + "\n"))
			sendLog(targetURL, line)
		}
	}()

	// stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(pipeStderr)
		for scanner.Scan() {
			line := scanner.Text()
			writer.Write([]byte(line + "\n"))
			sendLog(targetURL, line)
		}
	}()

	wg.Wait()
}
