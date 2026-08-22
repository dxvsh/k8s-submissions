package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

const (
	counterID  = 1
	listenAddr = ":3000"
)

var db *sql.DB

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := incrementPingCount()
	if err != nil {
		log.Printf("failed to increment ping count: %v", err)
		http.Error(w, "failed to update ping count", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "pong %d", count)
}

func pingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := getPingCount()
	if err != nil {
		log.Printf("failed to get ping count: %v", err)
		http.Error(w, "failed to get ping count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ping_count": %d}`, count)
}

func incrementPingCount() (int64, error) {
	var count int64
	err := db.QueryRow(`
		UPDATE ping_counts
		SET ping_count = ping_count + 1
		WHERE id = $1
		RETURNING ping_count
	`, counterID).Scan(&count)
	return count, err
}

func getPingCount() (int64, error) {
	var count int64
	err := db.QueryRow(`
		SELECT ping_count
		FROM ping_counts
		WHERE id = $1
	`, counterID).Scan(&count)
	return count, err
}

func initDB() (*sql.DB, error) {
	database, err := sql.Open("postgres", postgresURL())
	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}

	if _, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS ping_counts (
			id integer PRIMARY KEY,
			ping_count bigint NOT NULL
		)
	`); err != nil {
		database.Close()
		return nil, err
	}

	if _, err := database.Exec(`
		INSERT INTO ping_counts (id, ping_count)
		VALUES ($1, 0)
		ON CONFLICT (id) DO NOTHING
	`, counterID); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func postgresURL() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	databaseName := os.Getenv("POSTGRES_DB")
	sslMode := os.Getenv("POSTGRES_SSLMODE")

	hostPort := host
	if _, _, err := net.SplitHostPort(host); err != nil {
		hostPort = net.JoinHostPort(host, port)
	}

	postgresURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   hostPort,
		Path:   "/" + databaseName,
	}

	query := postgresURL.Query()
	query.Set("sslmode", sslMode)
	postgresURL.RawQuery = query.Encode()

	return postgresURL.String()
}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("failed to initialize postgres: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/pingpong", pingPongHandler)
	http.HandleFunc("/pings", pingsHandler)

	log.Println("listening on", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
