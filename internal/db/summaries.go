package db

import (
	"fmt"
	"time"
)

// Summary represents a session-end summary.
type Summary struct {
	ID        int64
	SessionID string
	Text      string
	Project   string
	CreatedAt time.Time
}

// InsertSummary stores a new session summary.
func (db *DB) InsertSummary(s *Summary) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO summaries (session_id, text) VALUES (?, ?)`,
		s.SessionID, s.Text,
	)
	if err != nil {
		return 0, fmt.Errorf("insert summary: %w", err)
	}
	return res.LastInsertId()
}

// RecentSummaries returns the N most recent summaries with project info from the linked session.
func (db *DB) RecentSummaries(limit int) ([]*Summary, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := db.conn.Query(
		`SELECT s.id, s.session_id, s.text, s.created_at, COALESCE(sess.project, '')
		 FROM summaries s
		 LEFT JOIN sessions sess ON s.session_id = sess.id
		 ORDER BY s.created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent summaries: %w", err)
	}
	defer rows.Close()

	var results []*Summary
	for rows.Next() {
		s := &Summary{}
		var createdAt string
		if err := rows.Scan(&s.ID, &s.SessionID, &s.Text, &createdAt, &s.Project); err != nil {
			return nil, fmt.Errorf("scan summary: %w", err)
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, s)
	}
	return results, rows.Err()
}
