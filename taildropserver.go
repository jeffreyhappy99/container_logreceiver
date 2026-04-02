package main

import (
	"io"
	"log"
	"net/http"
)

var queue = make(chan string, 1)
var drop int

func printLog() {
	go func() {
		for msg := range queue {
			log.Printf("%q", msg)
		}
	}()
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	// 限定只接受 POST
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// 記得最後關掉 body
	defer r.Body.Close()

	// 把整個 body 讀出來（就是你那個 line + "\n"）
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	line := string(data)

	select {
	case queue <- line:

	default:
		drop++
		log.Printf("[TAIL-DROP] buffer full drop log=%q (dropped=%d)\n", line, drop)
	}
}

func main() {
	printLog()
	http.HandleFunc("/logsample", logHandler)
	addr := ":8080"
	log.Printf("log receiver listening on %s ...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
