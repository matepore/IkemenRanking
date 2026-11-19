package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matepore/IkemenRanking/internal/tournament"
	_ "modernc.org/sqlite"
)

const databaseSchemaVersion = 1

// Database stores the latest completed result for each unordered character pair.
// The SQLite driver is pure Go, so SQLite is embedded in the application binary.
type Database struct {
	db *sql.DB
}

// OpenDatabase opens or creates a persistent IkemenRanking database.
func OpenDatabase(filePath string) (*Database, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("database path is required")
	}
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not resolve database path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return nil, fmt.Errorf("could not create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", absolutePath)
	if err != nil {
		return nil, fmt.Errorf("could not open SQLite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	database := &Database{db: db}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("could not connect to SQLite database: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("could not configure SQLite busy timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("could not enable SQLite write-ahead logging: %w", err)
	}
	if err := database.initialize(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return database, nil
}

// OpenResultsDatabase opens the persistent database inside the app's results directory.
func OpenResultsDatabase(resultsDir string) (*Database, error) {
	return OpenDatabase(filepath.Join(resultsDir, "ikemen-ranking.db"))
}

// Close releases the underlying SQLite connection.
func (database *Database) Close() error {
	if database == nil || database.db == nil {
		return nil
	}
	return database.db.Close()
}

func (database *Database) initialize(ctx context.Context) error {
	var version int
	if err := database.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("could not read database schema version: %w", err)
	}
	if version > databaseSchemaVersion {
		return fmt.Errorf("database schema version %d is newer than supported version %d", version, databaseSchemaVersion)
	}
	if version == databaseSchemaVersion {
		return nil
	}
	tx, err := database.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start database schema transaction: %w", err)
	}
	defer tx.Rollback()
	schema := "CREATE TABLE IF NOT EXISTS fights (" +
		"player1_key TEXT NOT NULL, player2_key TEXT NOT NULL, " +
		"player1 TEXT NOT NULL, player2 TEXT NOT NULL, " +
		"player1_wins INTEGER NOT NULL CHECK (player1_wins >= 0), " +
		"player2_wins INTEGER NOT NULL CHECK (player2_wins >= 0), " +
		"draws INTEGER NOT NULL CHECK (draws >= 0), " +
		"winner TEXT NOT NULL DEFAULT '', " +
		"is_draw INTEGER NOT NULL CHECK (is_draw IN (0, 1)), " +
		"stage TEXT NOT NULL DEFAULT '', updated_at TEXT NOT NULL, " +
		"PRIMARY KEY (player1_key, player2_key), CHECK (player1_key < player2_key))"
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("could not create fights table: %w", err)
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", databaseSchemaVersion)); err != nil {
		return fmt.Errorf("could not set database schema version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit database schema: %w", err)
	}
	return nil
}

// UpsertFight inserts a completed fight or replaces the current result for that
// unordered pair. A reversed input order is normalized, including its score.
func (database *Database) UpsertFight(ctx context.Context, record tournament.MatchRecord) error {
	player1, player2 := strings.TrimSpace(record.Match.Player1), strings.TrimSpace(record.Match.Player2)
	if player1 == "" || player2 == "" {
		return fmt.Errorf("fight participant names cannot be empty")
	}
	key1, key2 := strings.ToLower(player1), strings.ToLower(player2)
	if key1 == key2 {
		return fmt.Errorf("a character cannot fight itself")
	}
	result := record.Result
	if result.Player1Wins < 0 || result.Player2Wins < 0 || result.Draws < 0 {
		return fmt.Errorf("fight result counters cannot be negative")
	}
	if result.IsDraw {
		if result.Winner != "" {
			return fmt.Errorf("a draw cannot declare a winner")
		}
	} else if !strings.EqualFold(result.Winner, player1) && !strings.EqualFold(result.Winner, player2) {
		return fmt.Errorf("fight winner must be one of its participants")
	}
	if key1 > key2 {
		player1, player2 = player2, player1
		key1, key2 = key2, key1
		result.Player1Wins, result.Player2Wins = result.Player2Wins, result.Player1Wins
	}
	query := "INSERT INTO fights (player1_key, player2_key, player1, player2, " +
		"player1_wins, player2_wins, draws, winner, is_draw, stage, updated_at) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) " +
		"ON CONFLICT(player1_key, player2_key) DO UPDATE SET " +
		"player1 = excluded.player1, player2 = excluded.player2, " +
		"player1_wins = excluded.player1_wins, player2_wins = excluded.player2_wins, " +
		"draws = excluded.draws, winner = excluded.winner, is_draw = excluded.is_draw, " +
		"stage = excluded.stage, updated_at = excluded.updated_at"
	_, err := database.db.ExecContext(ctx, query,
		key1, key2, player1, player2,
		result.Player1Wins, result.Player2Wins, result.Draws,
		result.Winner, boolToInt(result.IsDraw), result.Stage,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("could not save completed fight: %w", err)
	}
	return nil
}

// ListFightsForParticipants returns the latest saved result for every pair
// contained in participants, in stable alphabetical order.
func (database *Database) ListFightsForParticipants(ctx context.Context, participants []string) ([]tournament.MatchRecord, error) {
	selected := make(map[string]struct{}, len(participants))
	for _, participant := range participants {
		name := strings.TrimSpace(participant)
		if name == "" {
			return nil, fmt.Errorf("participant names cannot be empty")
		}
		key := strings.ToLower(name)
		if _, exists := selected[key]; exists {
			return nil, fmt.Errorf("duplicate participant: %s", participant)
		}
		selected[key] = struct{}{}
	}
	rows, err := database.db.QueryContext(ctx,
		"SELECT player1, player2, player1_wins, player2_wins, draws, winner, is_draw, stage " +
			"FROM fights ORDER BY player1_key, player2_key")
	if err != nil {
		return nil, fmt.Errorf("could not query saved fights: %w", err)
	}
	defer rows.Close()
	records := make([]tournament.MatchRecord, 0)
	for rows.Next() {
		var record tournament.MatchRecord
		var isDraw int
		if err := rows.Scan(
			&record.Match.Player1, &record.Match.Player2,
			&record.Result.Player1Wins, &record.Result.Player2Wins,
			&record.Result.Draws, &record.Result.Winner, &isDraw, &record.Result.Stage,
		); err != nil {
			return nil, fmt.Errorf("could not read saved fight: %w", err)
		}
		_, p1Selected := selected[strings.ToLower(record.Match.Player1)]
		_, p2Selected := selected[strings.ToLower(record.Match.Player2)]
		if !p1Selected || !p2Selected {
			continue
		}
		record.Result.IsDraw = isDraw != 0
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not finish reading saved fights: %w", err)
	}
	for i := range records {
		records[i].Index = i + 1
	}
	return records, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
