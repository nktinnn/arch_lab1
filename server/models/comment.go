package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	TicketID  int       `json:"ticket_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentStore struct{ DB *sql.DB }

func NewCommentStore(db *sql.DB) *CommentStore { return &CommentStore{DB: db} }

func (s *CommentStore) Create(ticketID, userID int, content string) (*Comment, error) {
	c := &Comment{}
	err := s.DB.QueryRow(
		`INSERT INTO comments (ticket_id, user_id, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, ticket_id, user_id, content, created_at`,
		ticketID, userID, content,
	).Scan(&c.ID, &c.TicketID, &c.UserID, &c.Content, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	s.DB.QueryRow(`SELECT username FROM users WHERE id=$1`, c.UserID).Scan(&c.Username)
	return c, nil
}

func (s *CommentStore) ListByTicket(ticketID int) ([]*Comment, error) {
	rows, err := s.DB.Query(
		`SELECT c.id, c.ticket_id, c.user_id, u.username, c.content, c.created_at
		 FROM comments c
		 JOIN users u ON u.id = c.user_id
		 WHERE c.ticket_id=$1
		 ORDER BY c.created_at ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.TicketID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (s *CommentStore) Delete(id int) error {
	_, err := s.DB.Exec(`DELETE FROM comments WHERE id=$1`, id)
	return err
}
