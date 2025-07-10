package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // SQLite driver
)

type Metric struct {
	ID           int    `json:"id"`
	EventTime    string `json:"event_time"`
	Note         string `json:"note"`
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN environment variable not set")
	}

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS metrics(
				id INT AUTO_INCREMENT PRIMARY KEY,
				event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				note TEXT DEFAULT NULL,
				category_id INT NOT NULL
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

			CREATE TABLE IF NOT EXISTS categories(
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100) COLLATE utf8mb4_bin DEFAULT NULL
			) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Set up HTTP routes
	http.HandleFunc("/api/metrics", getMetrics)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Start server
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Get metrics from the database
func getMetrics(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT metrics.id, metrics.event_time, metrics.note, metrics.category_id, categories.name 
		FROM metrics 
		JOIN categories ON metrics.category_id = categories.id
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metric Metric
		if err := rows.Scan(&metric.ID, &metric.EventTime, &metric.Note, &metric.CategoryID, &metric.CategoryName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics = append(metrics, metric)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
