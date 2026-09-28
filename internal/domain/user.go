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
	RoleID           uuid.UUID  `db:"role_id" json:"role_id"`
	Avatar           string     `json:"avatar"`
	GoogleID         *string    `db:"google_id" json:"google_id,omitempty"`
	ProfilePicture   *string    `db:"profile_picture" json:"profile_picture,omitempty"`
	ResetToken       *string    `db:"reset_token" json:"-"`
	ResetTokenExpiry *time.Time `db:"reset_token_expiry" json:"-"`
	Status           string     `db:"status" json:"status"`
	DeletedAt        *time.Time `db:"deleted_at" json:"-"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}
