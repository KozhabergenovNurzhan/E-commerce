package models

import "time"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleManager  Role = "manager"
	RoleSeller   Role = "seller"
	RoleAdmin    Role = "admin"
)

type UserRecord struct {
	ID           int64      `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	FirstName    string     `db:"first_name"`
	LastName     string     `db:"last_name"`
	Role         Role       `db:"role"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

type Register struct {
	Email     string `json:"email"      binding:"required,email,max=255"`
	Password  string `json:"password"   binding:"required,min=8,max=72"`
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name"  binding:"required,max=100"`
}

type Login struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *UserRecord) ToResponse() *User {
	return &User{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

type Auth struct {
	User   *User       `json:"user"`
	Tokens *AuthTokens `json:"tokens"`
}
