package helpers

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func OpenDB(url string) *sql.DB {
	db, err := sql.Open("pgx", url)
	if err != nil {
		panic(err)
	}
	return db
}

func CleanupDB(db *sql.DB) {
	db.Exec("TRUNCATE courses CASCADE")
	db.Exec("TRUNCATE outbox_events CASCADE")
}

func CourseExists(db *sql.DB, id string) bool {
	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id=$1)", id).Scan(&exists)
	return exists
}

// OutboxEventExists checks if an outbox event with given aggregateID and eventType exists
func OutboxEventExists(db *sql.DB, aggregateID, eventType string) bool {
	var exists bool
	db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM outbox_events WHERE aggregate_id=$1 AND type=$2)",
		aggregateID, eventType,
	).Scan(&exists)
	return exists
}

// GetOutboxEventStatus returns the status of an outbox event
func GetOutboxEventStatus(db *sql.DB, aggregateID, eventType string) (string, error) {
	var status string
	err := db.QueryRow(
		"SELECT status FROM outbox_events WHERE aggregate_id=$1 AND type=$2 ORDER BY created_at DESC LIMIT 1",
		aggregateID, eventType,
	).Scan(&status)
	return status, err
}

// GetOutboxEventCount returns the count of outbox events with given aggregateID and eventType
func GetOutboxEventCount(db *sql.DB, aggregateID, eventType string) int {
	var count int
	db.QueryRow(
		"SELECT COUNT(*) FROM outbox_events WHERE aggregate_id=$1 AND type=$2",
		aggregateID, eventType,
	).Scan(&count)
	return count
}

// GetOutboxEventsWithStatus returns count of events with specific status
func GetOutboxEventsWithStatus(db *sql.DB, aggregateID, eventType, status string) int {
	var count int
	db.QueryRow(
		"SELECT COUNT(*) FROM outbox_events WHERE aggregate_id=$1 AND type=$2 AND status=$3",
		aggregateID, eventType, status,
	).Scan(&count)
	return count
}
