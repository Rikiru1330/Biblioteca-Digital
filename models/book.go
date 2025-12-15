package models

import "time"

type Book struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" binding:"required" db:"title"`
	Author      string    `json:"author" binding:"required" db:"author"`
	ISBN        string    `json:"isbn" binding:"required" db:"isbn"`
	Published   int       `json:"published" db:"published"`
	Genre       string    `json:"genre" db:"genre"`
	Description string    `json:"description" db:"description"`
	Available   bool      `json:"available" db:"available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Loan struct {
	ID         string     `json:"id" db:"id"`
	BookID     string     `json:"book_id" binding:"required" db:"book_id"` // ¡IMPORTANTE: db:"book_id"!
	User       string     `json:"user" binding:"required" db:"user"`
	LoanDate   time.Time  `json:"loan_date" db:"loan_date"`
	ReturnDate *time.Time `json:"return_date,omitempty" db:"return_date"`
	Returned   bool       `json:"returned" db:"returned"`
}

type CreateBookRequest struct {
	Title       string `json:"title" binding:"required"`
	Author      string `json:"author" binding:"required"`
	ISBN        string `json:"isbn" binding:"required"`
	Published   int    `json:"published"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
}

type UpdateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	ISBN        string `json:"isbn"`
	Published   int    `json:"published"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
	Available   *bool  `json:"available"`
}

type LoanRequest struct {
	BookID string `json:"book_id" binding:"required"`
	User   string `json:"user" binding:"required"`
}

// NOTA: Eliminamos la importación de uuid de aquí
// porque solo se usa en storage/memory_store.go
