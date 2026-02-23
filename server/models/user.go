package models

import (
	"database/sql"
	"time"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleUser     Role = "user"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStore struct{ DB *sql.DB }

func NewUserStore(db *sql.DB) *UserStore { return &UserStore{DB: db} }

func (s *UserStore) Create(username, email, passwordHash string, role Role) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow(
		`INSERT INTO users (username, email, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, email, role, created_at`,
		username, email, passwordHash, role,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	return u, err
}

func (s *UserStore) GetByEmail(email string) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserStore) GetByID(id int) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow(
		`SELECT id, username, email, role, created_at FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserStore) List() ([]*User, error) {
	rows, err := s.DB.Query(
		`SELECT id, username, email, role, created_at FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *UserStore) UpdateRole(id int, role Role) error {
	_, err := s.DB.Exec(`UPDATE users SET role=$1 WHERE id=$2`, role, id)
	return err
}

func (s *UserStore) Delete(id int) error {
	_, err := s.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}
