package storage

import (
	"fmt"
	"library-api/models"
)

// Errores comunes
var (
	ErrBookNotFound       = fmt.Errorf("book not found")
	ErrBookNotAvailable   = fmt.Errorf("book not available")
	ErrLoanNotFound       = fmt.Errorf("loan not found")
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrUserAlreadyExists  = fmt.Errorf("user already exists")
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
)

type Store interface {
	// ========== MÉTODOS PARA USUARIOS ==========
	CreateUser(user models.User) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(id string, user models.User) (*models.User, error)
	DeleteUser(id string) error

	// ========== MÉTODOS PARA LIBROS ==========
	CreateBook(book models.Book) (*models.Book, error)
	GetBooks() ([]models.Book, error)
	GetBookByID(id string) (*models.Book, error)
	UpdateBook(id string, book models.Book) (*models.Book, error)
	DeleteBook(id string) error
	SearchBooks(title, author, genre string, available *bool) ([]models.Book, error)

	// ========== MÉTODOS PARA PRÉSTAMOS ==========
	CreateLoan(loan models.Loan) (*models.Loan, error)
	ReturnBook(loanID string) error
	GetLoans() ([]models.Loan, error)
	GetActiveLoans() ([]models.Loan, error)
	GetLoanByID(id string) (*models.Loan, error)
}
