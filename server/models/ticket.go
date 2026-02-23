package models

import (
	"database/sql"
	"time"
)

type TicketStatus string
type TicketPriority string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"

	PriorityLow      TicketPriority = "low"
	PriorityMedium   TicketPriority = "medium"
	PriorityHigh     TicketPriority = "high"
	PriorityCritical TicketPriority = "critical"
)

type Ticket struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`
	AuthorID    int            `json:"author_id"`
	AuthorName  string         `json:"author_name"`
	AssignedTo  *int           `json:"assigned_to"`
	AssigneeName *string       `json:"assignee_name"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type TicketStore struct{ DB *sql.DB }

func NewTicketStore(db *sql.DB) *TicketStore { return &TicketStore{DB: db} }

func (s *TicketStore) Create(title, description string, priority TicketPriority, authorID int) (*Ticket, error) {
	t := &Ticket{}
	var assigneeName sql.NullString
	err := s.DB.QueryRow(
		`INSERT INTO tickets (title, description, priority, author_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, title, description, status, priority, author_id, assigned_to, created_at, updated_at`,
		title, description, priority, authorID,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AuthorID, &t.AssignedTo, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = assigneeName
	// fetch author name
	s.DB.QueryRow(`SELECT username FROM users WHERE id=$1`, t.AuthorID).Scan(&t.AuthorName)
	return t, nil
}

func (s *TicketStore) GetByID(id int) (*Ticket, error) {
	t := &Ticket{}
	var assigneeName sql.NullString
	err := s.DB.QueryRow(
		`SELECT t.id, t.title, t.description, t.status, t.priority,
		        t.author_id, u.username, t.assigned_to,
		        (SELECT username FROM users WHERE id=t.assigned_to),
		        t.created_at, t.updated_at
		 FROM tickets t
		 JOIN users u ON u.id = t.author_id
		 WHERE t.id=$1`,
		id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AuthorID, &t.AuthorName, &t.AssignedTo, &assigneeName,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if assigneeName.Valid {
		t.AssigneeName = &assigneeName.String
	}
	return t, nil
}

func (s *TicketStore) List(authorID *int) ([]*Ticket, error) {
	query := `SELECT t.id, t.title, t.description, t.status, t.priority,
	                 t.author_id, u.username, t.assigned_to,
	                 (SELECT username FROM users WHERE id=t.assigned_to),
	                 t.created_at, t.updated_at
	          FROM tickets t
	          JOIN users u ON u.id = t.author_id`
	args := []interface{}{}
	if authorID != nil {
		query += " WHERE t.author_id=$1"
		args = append(args, *authorID)
	}
	query += " ORDER BY t.created_at DESC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		t := &Ticket{}
		var assigneeName sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.AuthorID, &t.AuthorName, &t.AssignedTo, &assigneeName,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if assigneeName.Valid {
			t.AssigneeName = &assigneeName.String
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

func (s *TicketStore) UpdateStatus(id int, status TicketStatus) error {
	_, err := s.DB.Exec(`UPDATE tickets SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (s *TicketStore) UpdateAssignee(id int, assigneeID *int) error {
	_, err := s.DB.Exec(`UPDATE tickets SET assigned_to=$1 WHERE id=$2`, assigneeID, id)
	return err
}

func (s *TicketStore) Update(id int, title, description string, priority TicketPriority, status TicketStatus) error {
	_, err := s.DB.Exec(
		`UPDATE tickets SET title=$1, description=$2, priority=$3, status=$4 WHERE id=$5`,
		title, description, priority, status, id,
	)
	return err
}

func (s *TicketStore) Delete(id int) error {
	_, err := s.DB.Exec(`DELETE FROM tickets WHERE id=$1`, id)
	return err
}
