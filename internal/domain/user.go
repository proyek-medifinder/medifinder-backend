package domain

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	Email    string    `db:"email"`
	Password string    `db:"password"`
	RoleID   int       `db:"role_id"`
}
