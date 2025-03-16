package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user entity
type User struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	Email     string    `json:"email" bson:"email"`
	Username  string    `json:"username" bson:"username"`
	Password  string    `json:"-" bson:"password"` // Never expose password in JSON responses
	FirstName string    `json:"first_name" bson:"first_name"`
	LastName  string    `json:"last_name" bson:"last_name"`
	Role      string    `json:"role" bson:"role"`
	Status    string    `json:"status" bson:"status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// UserStatus enum
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBlocked  = "blocked"
)

// UserRole enum
const (
	UserRoleAdmin  = "admin"
	UserRoleUser   = "user"
	UserRoleMember = "member"
)

// NewUser creates a new user with default values
func NewUser(email, username, password, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Username:  username,
		Password:  password, // Note: Should be hashed before saving
		FirstName: firstName,
		LastName:  lastName,
		Role:      UserRoleUser,
		Status:    UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
