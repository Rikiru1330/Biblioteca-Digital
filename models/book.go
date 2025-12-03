package models

import (
	"time"
)

type Book struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" binding:"required"`
	Author      string    `json:"author" binding:"required"`
	ISBN        string    `json:"isbn" binding:"required"`
	Published   int       `json:"published"`
	Genre       string    `json:"genre"`
	Description string    `json:"description"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Loan struct {
	ID         string     `json:"id"`
	BookID     string     `json:"book_id" binding:"required"`
	User       string     `json:"user" binding:"required"`
	LoanDate   time.Time  `json:"loan_date"`
	ReturnDate *time.Time `json:"return_date,omitempty"`
	Returned   bool       `json:"returned"`
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