package models

import "time"

type Address struct {
	ID         int64     `db:"id"          json:"id"`
	UserID     int64     `db:"user_id"      json:"user_id"`
	FullName   string    `db:"full_name"    json:"full_name"`
	Phone      string    `db:"phone"        json:"phone"`
	Country    string    `db:"country"      json:"country"`
	City       string    `db:"city"         json:"city"`
	Street     string    `db:"street"       json:"street"`
	PostalCode string    `db:"postal_code"  json:"postal_code"`
	IsDefault  bool      `db:"is_default"   json:"is_default"`
	CreatedAt  time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"   json:"updated_at"`
}

type CreateAddress struct {
	FullName   string `json:"full_name"   binding:"required,max=200"`
	Phone      string `json:"phone"       binding:"required,max=20"`
	Country    string `json:"country"     binding:"required,max=100"`
	City       string `json:"city"        binding:"required,max=100"`
	Street     string `json:"street"      binding:"required,max=255"`
	PostalCode string `json:"postal_code" binding:"required,max=20"`
	IsDefault  bool   `json:"is_default"`
}

type UpdateAddress struct {
	FullName   string `json:"full_name"   binding:"required,max=200"`
	Phone      string `json:"phone"       binding:"required,max=20"`
	Country    string `json:"country"     binding:"required,max=100"`
	City       string `json:"city"        binding:"required,max=100"`
	Street     string `json:"street"      binding:"required,max=255"`
	PostalCode string `json:"postal_code" binding:"required,max=20"`
}
