package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

const pingsFile = "logs/pings.txt"

var (
	counter atomic.Uint64
	pingMu  sync.Mutex
)

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pingMu.Lock()
	count := counter.Add(1)
	if err := os.WriteFile(pingsFile, fmt.Appendf(nil, "%d\n", count), 0644); err != nil {
		pingMu.Unlock()
		log.Printf("failed to update %s: %v", pingsFile, err)
		http.Error(w, "failed to update ping count", http.StatusInternalServerError)
		return
	}
	pingMu.Unlock()

	fmt.Fprintf(w, "pong %d", count)
}

func main() {
	http.HandleFunc("/pingpong", pingPongHandler)

	log.Println("listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
