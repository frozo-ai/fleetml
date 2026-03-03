package offline

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fleetml/fleetml/agent/internal/health"
	_ "modernc.org/sqlite" // Pure-Go SQLite driver (no CGO)
)

// SQLiteStore implements MetricsStore using pure-Go SQLite for offline buffering.
// Uses modernc.org/sqlite to avoid CGO dependency for cross-compilation.
type SQLiteStore struct {
	db *sql.DB
	mu sync.Mutex
}

// NewSQLiteStore opens (or creates) a SQLite database at the given path.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS heartbeat_buffer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id TEXT NOT NULL,
			status TEXT NOT NULL,
			system_metrics TEXT,
			timestamp INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS command_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command_id TEXT NOT NULL UNIQUE,
			command_type TEXT NOT NULL,
			payload TEXT,
			received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			processed INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_heartbeat_timestamp ON heartbeat_buffer(timestamp);
		CREATE INDEX IF NOT EXISTS idx_command_processed ON command_queue(processed);
	`)
	return err
}

// SaveHeartbeat buffers a heartbeat record locally.
func (s *SQLiteStore) SaveHeartbeat(record *HeartbeatRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metricsJSON []byte
	if record.System != nil {
		var err error
		metricsJSON, err = json.Marshal(record.System)
		if err != nil {
			return fmt.Errorf("marshal metrics: %w", err)
		}
	}

	_, err := s.db.Exec(`
		INSERT INTO heartbeat_buffer (device_id, status, system_metrics, timestamp)
		VALUES (?, ?, ?, ?)`,
		record.DeviceID, record.Status, string(metricsJSON), record.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("insert heartbeat: %w", err)
	}

	return nil
}

// GetBufferedHeartbeats retrieves all buffered heartbeats ordered by timestamp.
func (s *SQLiteStore) GetBufferedHeartbeats() ([]*HeartbeatRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT device_id, status, system_metrics, timestamp
		FROM heartbeat_buffer
		ORDER BY timestamp ASC`)
	if err != nil {
		return nil, fmt.Errorf("query heartbeats: %w", err)
	}
	defer rows.Close()

	var records []*HeartbeatRecord
	for rows.Next() {
		var r HeartbeatRecord
		var metricsJSON sql.NullString
		if err := rows.Scan(&r.DeviceID, &r.Status, &metricsJSON, &r.Timestamp); err != nil {
			return nil, fmt.Errorf("scan heartbeat: %w", err)
		}

		if metricsJSON.Valid && metricsJSON.String != "" {
			var m health.SystemMetrics
			if err := json.Unmarshal([]byte(metricsJSON.String), &m); err == nil {
				r.System = &m
			}
		}

		records = append(records, &r)
	}

	return records, nil
}

// ClearBuffer removes all buffered heartbeats.
func (s *SQLiteStore) ClearBuffer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM heartbeat_buffer")
	if err != nil {
		return fmt.Errorf("clear buffer: %w", err)
	}
	return nil
}

// BufferCount returns the number of buffered heartbeats.
func (s *SQLiteStore) BufferCount() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM heartbeat_buffer").Scan(&count)
	return count, err
}

// SaveCommand queues a command for later processing.
func (s *SQLiteStore) SaveCommand(commandID, commandType string, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO command_queue (command_id, command_type, payload)
		VALUES (?, ?, ?)`,
		commandID, commandType, string(payload),
	)
	return err
}

// GetPendingCommands returns unprocessed commands.
func (s *SQLiteStore) GetPendingCommands() ([]struct {
	ID      string
	Type    string
	Payload []byte
}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT command_id, command_type, payload
		FROM command_queue
		WHERE processed = 0
		ORDER BY received_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []struct {
		ID      string
		Type    string
		Payload []byte
	}
	for rows.Next() {
		var cmd struct {
			ID      string
			Type    string
			Payload []byte
		}
		var payloadStr string
		if err := rows.Scan(&cmd.ID, &cmd.Type, &payloadStr); err != nil {
			return nil, err
		}
		cmd.Payload = []byte(payloadStr)
		commands = append(commands, cmd)
	}

	return commands, nil
}

// MarkCommandProcessed marks a command as processed.
func (s *SQLiteStore) MarkCommandProcessed(commandID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("UPDATE command_queue SET processed = 1 WHERE command_id = ?", commandID)
	return err
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
