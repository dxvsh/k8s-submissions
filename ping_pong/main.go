package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

var counter atomic.Uint64

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count := counter.Add(1)
	fmt.Fprintf(w, "pong %d", count)
}

func main() {
	http.HandleFunc("/pingpong", pingPongHandler)

	log.Println("listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
