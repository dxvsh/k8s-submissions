package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const maxTodoLength = 140

var requestLogger = log.New(os.Stdout, "", 0)

type todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type todoStore struct {
	mu     sync.RWMutex
	nextID int
	todos  []todo
}

func newTodoStore() *todoStore {
	return &todoStore{
		nextID: 1,
		todos:  make([]todo, 0),
	}
}

func (s *todoStore) all() []todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]todo, len(s.todos))
	copy(todos, s.todos)
	return todos
}

func (s *todoStore) add(text string) todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	newTodo := todo{ID: s.nextID, Text: text}
	s.nextID++
	s.todos = append(s.todos, newTodo)
	return newTodo
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4040"
	}

	store := newTodoStore()
	mux := http.NewServeMux()
	mux.Handle("/api/todos", todosHandler(store))

	log.Printf("Server started in port %s", port)
	handler := logRequests(enableCORS(mux))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func todosHandler(store *todoStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, store.all())
		case http.MethodPost:
			createTodo(w, r, store)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

func createTodo(w http.ResponseWriter, r *http.Request, store *todoStore) {
	var input struct {
		Text string `json:"text"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Request body must contain valid JSON")
		return
	}

	if err := ensureSingleJSONValue(decoder); err != nil {
		writeError(w, http.StatusBadRequest, "Request body must contain one JSON object")
		return
	}

	text := strings.TrimSpace(input.Text)
	if text == "" {
		writeError(w, http.StatusBadRequest, "Todo text cannot be empty")
		return
	}
	if utf8.RuneCountInString(text) > maxTodoLength {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Todo text cannot exceed %d characters", maxTodoLength))
		return
	}

	writeJSON(w, http.StatusCreated, store.add(text))
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request contains additional JSON values")
	}
	return nil
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		requestLogger.Printf(
			"[%s] %q %d",
			time.Now().Format(time.DateTime),
			r.Method+" "+r.URL.RequestURI(),
			status,
		)
	})
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
