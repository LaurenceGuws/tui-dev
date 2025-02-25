package storage

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Message represents a chat message stored in the database.
type Message struct {
	ID             int64
	ConversationID string
	Role           string
	Content        string
	Timestamp      time.Time
}

// Storage encapsulates the SQLite database.
type Storage struct {
	DB     *sql.DB
	Logger *log.Logger
}

// NewStorage opens (or creates) a SQLite database at dbPath and returns a Storage.
func NewStorage(dbPath string, logger *log.Logger) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Println("ERROR: Failed to open database:", err)
		return nil, err
	}
	s := &Storage{DB: db, Logger: logger}
	if err := s.Init(); err != nil {
		logger.Println("ERROR: Failed to initialize database:", err)
		return nil, err
	}
	logger.Println("INFO: Database initialized successfully at", dbPath)
	return s, nil
}

// Init creates the messages table if it does not exist.
// A new column "conversation_id" is added to group messages.
func (s *Storage) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id TEXT,
		role TEXT,
		content TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.DB.Exec(query)
	if err != nil {
		s.Logger.Println("ERROR: Failed to create messages table:", err)
	} else {
		s.Logger.Println("INFO: Messages table initialized (or already exists)")
	}
	return err
}

// AddMessage inserts a new message into the database for a given conversation.
func (s *Storage) AddMessage(conversationID, role, content string) error {
	query := `INSERT INTO messages(conversation_id, role, content) VALUES(?, ?, ?)`
	result, err := s.DB.Exec(query, conversationID, role, content)
	if err != nil {
		s.Logger.Println("ERROR: Failed to insert message:", err)
		return err
	}
	id, _ := result.LastInsertId()
	s.Logger.Printf("INFO: Message added (ID: %d, Conversation: %s, Role: %s, Content: %s)", id, conversationID, role, content)
	return nil
}

// GetMessagesForConversation retrieves all messages for a given conversation.
func (s *Storage) GetMessagesForConversation(conversationID string) ([]Message, error) {
	query := "SELECT id, conversation_id, role, content, timestamp FROM messages WHERE conversation_id = ? ORDER BY id"
	s.Logger.Println("INFO: Fetching messages for conversation:", conversationID)
	rows, err := s.DB.Query(query, conversationID)
	if err != nil {
		s.Logger.Println("ERROR: Failed to fetch messages:", err)
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var ts string
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &ts); err != nil {
			s.Logger.Println("ERROR: Failed to scan message row:", err)
			return nil, err
		}
		msg.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		messages = append(messages, msg)
	}
	s.Logger.Printf("INFO: Retrieved %d messages for conversation %s", len(messages), conversationID)
	return messages, nil
}

// GetConversations retrieves distinct conversation IDs.
func (s *Storage) GetConversations() ([]string, error) {
	query := "SELECT DISTINCT conversation_id FROM messages ORDER BY conversation_id"
	s.Logger.Println("INFO: Fetching distinct conversation IDs")
	rows, err := s.DB.Query(query)
	if err != nil {
		s.Logger.Println("ERROR: Failed to fetch conversation IDs:", err)
		return nil, err
	}
	defer rows.Close()

	var conversations []string
	for rows.Next() {
		var convID string
		if err := rows.Scan(&convID); err != nil {
			s.Logger.Println("ERROR: Failed to scan conversation ID:", err)
			return nil, err
		}
		conversations = append(conversations, convID)
	}
	s.Logger.Printf("INFO: Retrieved %d conversations", len(conversations))
	return conversations, nil
}

// DeleteMessage removes a message by ID.
func (s *Storage) DeleteMessage(id int64) error {
	query := "DELETE FROM messages WHERE id = ?"
	result, err := s.DB.Exec(query, id)
	if err != nil {
		s.Logger.Println("ERROR: Failed to delete message with ID", id, ":", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	s.Logger.Printf("INFO: Deleted message ID %d (Rows affected: %d)", id, rowsAffected)
	return nil
}

// UpdateMessage updates the content of a message by ID.
func (s *Storage) UpdateMessage(id int64, newContent string) error {
	query := "UPDATE messages SET content = ? WHERE id = ?"
	result, err := s.DB.Exec(query, newContent, id)
	if err != nil {
		s.Logger.Println("ERROR: Failed to update message with ID", id, ":", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	s.Logger.Printf("INFO: Updated message ID %d (Rows affected: %d, New Content: %s)", id, rowsAffected, newContent)
	return nil
}
