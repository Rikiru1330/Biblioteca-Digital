package storage

import "library-api/models"

type Store interface {
	// Métodos para libros
	CreateBook(book models.Book) (models.Book, error)
	GetBooks() ([]models.Book, error)
	GetBookByID(id string) (models.Book, error)
	UpdateBook(id string, book models.Book) (models.Book, error)
	DeleteBook(id string) error
	SearchBooks(title, author, genre string, available *bool) ([]models.Book, error)
	
	// Métodos para préstamos
	CreateLoan(loan models.Loan) (models.Loan, error)
	ReturnBook(loanID string) error
	GetLoans() ([]models.Loan, error)
	GetActiveLoans() ([]models.Loan, error)
	GetLoanByID(id string) (models.Loan, error)
}