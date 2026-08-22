package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const maxTodoLength = 140

var requestLogger = log.New(os.Stdout, "", 0)

type todo struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type todoRepository interface {
	all(context.Context) ([]todo, error)
	add(context.Context, string) (todo, error)
}

type postgresTodoRepository struct {
	db *sql.DB
}

func (r *postgresTodoRepository) all(ctx context.Context) ([]todo, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, text FROM todos ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("querying todos: %w", err)
	}
	defer rows.Close()

	todos := make([]todo, 0)
	for rows.Next() {
		var item todo
		if err := rows.Scan(&item.ID, &item.Text); err != nil {
			return nil, fmt.Errorf("scanning todo: %w", err)
		}
		todos = append(todos, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading todos: %w", err)
	}

	return todos, nil
}

func (r *postgresTodoRepository) add(ctx context.Context, text string) (todo, error) {
	item := todo{Text: text}
	if err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO todos (text) VALUES ($1) RETURNING id`,
		text,
	).Scan(&item.ID); err != nil {
		return todo{}, fmt.Errorf("inserting todo: %w", err)
	}

	return item, nil
}

func main() {
	port := envOrDefault("PORT", "4040")

	db, err := openDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	if err := initializeDatabase(db); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}

	repository := &postgresTodoRepository{db: db}
	mux := http.NewServeMux()
	mux.Handle("/api/todos", todosHandler(repository))

	log.Printf("Server started in port %s", port)
	handler := logRequests(enableCORS(mux))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func openDatabase() (*sql.DB, error) {
	host, err := requiredEnv("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}
	user, err := requiredEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}
	password, err := requiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}
	name, err := requiredEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, envOrDefault("POSTGRES_PORT", "5432")),
		Path:   name,
	}
	query := connectionURL.Query()
	query.Set("sslmode", envOrDefault("POSTGRES_SSLMODE", "disable"))
	connectionURL.RawQuery = query.Encode()

	db, err := sql.Open("pgx", connectionURL.String())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

func initializeDatabase(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS todos (
			id BIGSERIAL PRIMARY KEY,
			text VARCHAR(140) NOT NULL CHECK (char_length(text) BETWEEN 1 AND 140),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating todos table: %w", err)
	}
	return nil
}

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("environment variable %s is required", name)
	}
	return value, nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func todosHandler(repository todoRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listTodos(w, r, repository)
		case http.MethodPost:
			createTodo(w, r, repository)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

func listTodos(w http.ResponseWriter, r *http.Request, repository todoRepository) {
	todos, err := repository.all(r.Context())
	if err != nil {
		log.Printf("Failed to list todos: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to load todos")
		return
	}

	writeJSON(w, http.StatusOK, todos)
}

func createTodo(w http.ResponseWriter, r *http.Request, repository todoRepository) {
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

	item, err := repository.add(r.Context(), text)
	if err != nil {
		log.Printf("Failed to create todo: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create todo")
		return
	}

	writeJSON(w, http.StatusCreated, item)
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
