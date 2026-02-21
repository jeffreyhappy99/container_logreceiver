package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

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

	// 這裡你要幹嘛都可以：印到 stdout / 寫檔 / 丟到 queue / 解析 JSON...
	log.Printf("received log: %q", line)

	// 回個 200 OK
	fmt.Fprintln(w, "ok")
}

func main() {
	http.HandleFunc("/log", logHandler)

	addr := ":8080"
	log.Printf("log receiver listening on %s ...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
