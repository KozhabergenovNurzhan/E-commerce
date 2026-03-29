package models

import "time"

type Review struct {
	ID        int64     `db:"id"         json:"id"`
	ProductID int64     `db:"product_id" json:"product_id"`
	UserID    int64     `db:"user_id"    json:"user_id"`
	Rating    int       `db:"rating"     json:"rating"`
	Comment   string    `db:"comment"    json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateReview struct {
	Rating  int    `json:"rating"  binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=2000"`
}

type UpdateReview struct {
	Rating  int    `json:"rating"  binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=2000"`
}

type ProductRating struct {
	Average float64 `db:"average" json:"average"`
	Count   int     `db:"count"   json:"count"`
}
