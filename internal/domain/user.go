package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	Name             string     `db:"name" json:"name"`
	Email            string     `db:"email" json:"email"`
	Password         string     `db:"password" json:"-"`
	RoleID           int        `db:"role_id" json:"role_id"`
	GoogleID         *string    `db:"google_id" json:"google_id,omitempty"`
	ResetToken       *string    `db:"reset_token" json:"-"`
	ResetTokenExpiry *time.Time `db:"reset_token_expiry" json:"-"`
	Status           string     `db:"status" json:"status"`
	LastLoginAt      *time.Time `db:"last_login_at" json:"last_login_at"`
	DeletedAt        *time.Time `db:"deleted_at" json:"-"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}
