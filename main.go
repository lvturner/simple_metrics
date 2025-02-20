package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type Metric struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

var db *sql.DB

func main() {
	// Initialize database
	var err error
	db, err = sql.Open("sqlite3", "./metrics.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS metrics (name TEXT PRIMARY KEY, count INTEGER)`)
	if err != nil {
		log.Fatal(err)
	}

	// Seed with some data (optional)
	seedDatabase()

	// Set up HTTP routes
	http.HandleFunc("/api/metrics", getMetrics)
	http.HandleFunc("/api/add", addMetric)
	http.HandleFunc("/api/addMetricEntry", addMetricEntry)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Start server
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Get metrics from the database
func getMetrics(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT name, total_count FROM metrics")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metric Metric
		if err := rows.Scan(&metric.Name, &metric.Count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics = append(metrics, metric)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// Add a new metric
func addMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var metric Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO metrics (name) VALUES (?)", metric.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func addMetricEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var entry struct {
		MetricName string `json:"name"`
		Value      int    `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get metric ID
	var metricID int
	err = tx.QueryRow("SELECT id FROM metrics WHERE name = ?", entry.MetricName).Scan(&metricID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	// Insert new time-series entry
	_, err = tx.Exec("INSERT INTO metric_entries (metric_id, value) VALUES (?, ?)", metricID, entry.Value)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update total count
	_, err = tx.Exec("UPDATE metrics SET total_count = total_count + ? WHERE id = ?", entry.Value, metricID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusCreated)
}

// Seed the database with some data
func seedDatabase() {
	_, _ = db.Exec("INSERT OR IGNORE INTO metrics (name, count) VALUES ('Metric A', 10), ('Metric B', 20), ('Metric C', 30)")
}
