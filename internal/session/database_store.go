package session

import (
	"database/sql"
	"encoding/json"
	"fmt"

	providerpkg "agent-harness/internal/provider"

	_ "modernc.org/sqlite"
)

type DatabaseSessionStore struct {
	db *sql.DB
}

func NewDatabaseSessionStore(path string) (*DatabaseSessionStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &DatabaseSessionStore{
		db: db,
	}

	if err := store.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (store *DatabaseSessionStore) initialize() error {
	if _, err := store.db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return err
	}

	_, err := store.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			metadata TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			tool_calls TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		);
	`)

	return err
}

func (store *DatabaseSessionStore) Set(session *Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return err
	}

	transaction, err := store.db.Begin()
	if err != nil {
		return err
	}

	defer transaction.Rollback()

	_, err = transaction.Exec(`
		INSERT INTO sessions (
			id,
			created_at,
			updated_at,
			metadata
		)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			updated_at = excluded.updated_at,
			metadata = excluded.metadata
	`,
		session.ID,
		session.CreatedAt,
		session.UpdatedAt,
		string(metadata),
	)
	if err != nil {
		return err
	}

	if _, err := transaction.Exec(
		`DELETE FROM messages WHERE session_id = ?`,
		session.ID,
	); err != nil {
		return err
	}

	for _, message := range session.Messages {
		toolCalls, err := json.Marshal(message.ToolCalls)
		if err != nil {
			return err
		}

		_, err = transaction.Exec(`
			INSERT INTO messages (
				session_id,
				role,
				content,
				tool_calls
			)
			VALUES (?, ?, ?, ?)
		`,
			session.ID,
			message.Role,
			message.Content,
			string(toolCalls),
		)
		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}

func (store *DatabaseSessionStore) Get(id string) (*Session, error) {
	var session Session
	var metadata string

	err := store.db.QueryRow(`
		SELECT id, created_at, updated_at, metadata
		FROM sessions
		WHERE id = ?
	`, id).Scan(
		&session.ID,
		&session.CreatedAt,
		&session.UpdatedAt,
		&metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", id)
		}

		return nil, err
	}

	if err := json.Unmarshal([]byte(metadata), &session.Metadata); err != nil {
		return nil, err
	}

	session.Messages = make([]providerpkg.Message, 0)

	rows, err := store.db.Query(`
		SELECT role, content, tool_calls
		FROM messages
		WHERE session_id = ?
		ORDER BY id
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message providerpkg.Message
		var toolCalls string

		if err := rows.Scan(
			&message.Role,
			&message.Content,
			&toolCalls,
		); err != nil {
			return nil, err
		}

		if toolCalls != "" {
			if err := json.Unmarshal(
				[]byte(toolCalls),
				&message.ToolCalls,
			); err != nil {
				return nil, err
			}
		}

		session.Messages = append(session.Messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &session, nil
}

func (store *DatabaseSessionStore) Delete(id string) error {
	result, err := store.db.Exec(
		`DELETE FROM sessions WHERE id = ?`,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

func (store *DatabaseSessionStore) List() []*Session {
	rows, err := store.db.Query(`
		SELECT id
		FROM sessions
		ORDER BY created_at
	`)
	if err != nil {
		return []*Session{}
	}
	defer rows.Close()

	sessions := make([]*Session, 0)

	for rows.Next() {
		var id string

		if err := rows.Scan(&id); err != nil {
			return []*Session{}
		}

		session, err := store.Get(id)
		if err != nil {
			return []*Session{}
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return []*Session{}
	}

	return sessions
}
