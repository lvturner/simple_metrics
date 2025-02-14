-- Table to store metadata about metrics
CREATE TABLE metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique ID for each metric
    name TEXT NOT NULL UNIQUE,            -- Metric name (e.g., "Metric A")
    total_count INTEGER DEFAULT 0         -- Current total count (computed from metric_entries)
);

-- Table to store individual time-series entries for each metric
CREATE TABLE metric_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique ID for each entry
    metric_id INTEGER NOT NULL,           -- Foreign key to the metrics table
    value INTEGER NOT NULL,               -- Increment or decrement for the metric
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP, -- Time of the entry
    FOREIGN KEY (metric_id) REFERENCES metrics (id) ON DELETE CASCADE
);
