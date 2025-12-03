package storage

import (
	"database/sql"
	"fmt"
	"time"
	"library-api/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sqlx.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}
	
	return &SQLiteStore{db: db}, nil
}

func createTables(db *sqlx.DB) error {
	// Tabla de libros
	booksTable := `
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		isbn TEXT UNIQUE NOT NULL,
		published INTEGER,
		genre TEXT,
		description TEXT,
		available BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	
	// Tabla de préstamos
	loansTable := `
	CREATE TABLE IF NOT EXISTS loans (
		id TEXT PRIMARY KEY,
		book_id TEXT NOT NULL,
		user TEXT NOT NULL,
		loan_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		return_date TIMESTAMP,
		returned BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE CASCADE
	);
	`
	
	// Crear índices para búsquedas rápidas
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);",
		"CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);",
		"CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);",
		"CREATE INDEX IF NOT EXISTS idx_books_available ON books(available);",
		"CREATE INDEX IF NOT EXISTS idx_loans_book_id ON loans(book_id);",
		"CREATE INDEX IF NOT EXISTS idx_loans_returned ON loans(returned);",
	}
	
	if _, err := db.Exec(booksTable); err != nil {
		return err
	}
	
	if _, err := db.Exec(loansTable); err != nil {
		return err
	}
	
	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return err
		}
	}
	
	return nil
}

// CreateBook implementación
func (s *SQLiteStore) CreateBook(book models.Book) (models.Book, error) {
	book.ID = uuid.New().String()
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	book.Available = true
	
	query := `INSERT INTO books (id, title, author, isbn, published, genre, description, available, created_at, updated_at) 
	          VALUES (:id, :title, :author, :isbn, :published, :genre, :description, :available, :created_at, :updated_at)`
	
	_, err := s.db.NamedExec(query, book)
	if err != nil {
		return models.Book{}, fmt.Errorf("error creating book: %w", err)
	}
	
	return book, nil
}

// GetBooks implementación
func (s *SQLiteStore) GetBooks() ([]models.Book, error) {
	var books []models.Book
	query := `SELECT * FROM books ORDER BY title`
	
	err := s.db.Select(&books, query)
	if err != nil {
		return nil, fmt.Errorf("error getting books: %w", err)
	}
	
	return books, nil
}

// GetBookByID implementación
func (s *SQLiteStore) GetBookByID(id string) (models.Book, error) {
	var book models.Book
	query := `SELECT * FROM books WHERE id = ?`
	
	err := s.db.Get(&book, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Book{}, ErrBookNotFound
		}
		return models.Book{}, fmt.Errorf("error getting book: %w", err)
	}
	
	return book, nil
}

// UpdateBook implementación
func (s *SQLiteStore) UpdateBook(id string, updatedBook models.Book) (models.Book, error) {
	// Obtener libro existente
	existingBook, err := s.GetBookByID(id)
	if err != nil {
		return models.Book{}, err
	}
	
	// Mantener valores que no se actualizan
	updatedBook.ID = id
	updatedBook.CreatedAt = existingBook.CreatedAt
	updatedBook.UpdatedAt = time.Now()
	
	query := `UPDATE books SET 
		title = :title, 
		author = :author, 
		isbn = :isbn, 
		published = :published, 
		genre = :genre, 
		description = :description, 
		available = :available,
		updated_at = :updated_at
		WHERE id = :id`
	
	_, err = s.db.NamedExec(query, updatedBook)
	if err != nil {
		return models.Book{}, fmt.Errorf("error updating book: %w", err)
	}
	
	return updatedBook, nil
}

// DeleteBook implementación
func (s *SQLiteStore) DeleteBook(id string) error {
	query := `DELETE FROM books WHERE id = ?`
	
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting book: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrBookNotFound
	}
	
	return nil
}

// SearchBooks implementación
func (s *SQLiteStore) SearchBooks(title, author, genre string, available *bool) ([]models.Book, error) {
	var books []models.Book
	query := `SELECT * FROM books WHERE 1=1`
	args := []interface{}{}
	
	if title != "" {
		query += ` AND title LIKE ?`
		args = append(args, "%"+title+"%")
	}
	
	if author != "" {
		query += ` AND author LIKE ?`
		args = append(args, "%"+author+"%")
	}
	
	if genre != "" {
		query += ` AND genre LIKE ?`
		args = append(args, "%"+genre+"%")
	}
	
	if available != nil {
		query += ` AND available = ?`
		args = append(args, *available)
	}
	
	query += ` ORDER BY title`
	
	err := s.db.Select(&books, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching books: %w", err)
	}
	
	return books, nil
}

// CreateLoan implementación
func (s *SQLiteStore) CreateLoan(loan models.Loan) (models.Loan, error) {
	// Verificar que el libro existe y está disponible
	var available bool
	checkQuery := `SELECT available FROM books WHERE id = ?`
	err := s.db.Get(&available, checkQuery, loan.BookID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Loan{}, ErrBookNotFound
		}
		return models.Loan{}, fmt.Errorf("error checking book availability: %w", err)
	}
	
	if !available {
		return models.Loan{}, ErrBookNotAvailable
	}
	
	// Crear el préstamo
	loan.ID = uuid.New().String()
	loan.LoanDate = time.Now()
	loan.Returned = false
	
	tx, err := s.db.Beginx()
	if err != nil {
		return models.Loan{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Insertar préstamo
	loanQuery := `INSERT INTO loans (id, book_id, user, loan_date, returned) 
	              VALUES (:id, :book_id, :user, :loan_date, :returned)`
	
	_, err = tx.NamedExec(loanQuery, loan)
	if err != nil {
		return models.Loan{}, fmt.Errorf("error creating loan: %w", err)
	}
	
	// Actualizar libro como no disponible
	updateBookQuery := `UPDATE books SET available = FALSE, updated_at = ? WHERE id = ?`
	_, err = tx.Exec(updateBookQuery, time.Now(), loan.BookID)
	if err != nil {
		return models.Loan{}, fmt.Errorf("error updating book status: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		return models.Loan{}, fmt.Errorf("error committing transaction: %w", err)
	}
	
	return loan, nil
}

// ReturnBook implementación
func (s *SQLiteStore) ReturnBook(loanID string) error {
	// Verificar que el préstamo existe
	var loan models.Loan
	getLoanQuery := `SELECT * FROM loans WHERE id = ?`
	err := s.db.Get(&loan, getLoanQuery, loanID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrLoanNotFound
		}
		return fmt.Errorf("error getting loan: %w", err)
	}
	
	if loan.Returned {
		return nil // Ya está devuelto
	}
	
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Actualizar préstamo como devuelto
	now := time.Now()
	returnLoanQuery := `UPDATE loans SET returned = TRUE, return_date = ? WHERE id = ?`
	_, err = tx.Exec(returnLoanQuery, now, loanID)
	if err != nil {
		return fmt.Errorf("error updating loan: %w", err)
	}
	
	// Actualizar libro como disponible
	updateBookQuery := `UPDATE books SET available = TRUE, updated_at = ? WHERE id = ?`
	_, err = tx.Exec(updateBookQuery, now, loan.BookID)
	if err != nil {
		return fmt.Errorf("error updating book: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	
	return nil
}

// GetLoans implementación
func (s *SQLiteStore) GetLoans() ([]models.Loan, error) {
	var loans []models.Loan
	query := `SELECT * FROM loans ORDER BY loan_date DESC`
	
	err := s.db.Select(&loans, query)
	if err != nil {
		return nil, fmt.Errorf("error getting loans: %w", err)
	}
	
	return loans, nil
}

// GetActiveLoans implementación
func (s *SQLiteStore) GetActiveLoans() ([]models.Loan, error) {
	var loans []models.Loan
	query := `SELECT * FROM loans WHERE returned = FALSE ORDER BY loan_date DESC`
	
	err := s.db.Select(&loans, query)
	if err != nil {
		return nil, fmt.Errorf("error getting active loans: %w", err)
	}
	
	return loans, nil
}

// GetLoanByID implementación
func (s *SQLiteStore) GetLoanByID(id string) (models.Loan, error) {
	var loan models.Loan
	query := `SELECT * FROM loans WHERE id = ?`
	
	err := s.db.Get(&loan, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Loan{}, ErrLoanNotFound
		}
		return models.Loan{}, fmt.Errorf("error getting loan: %w", err)
	}
	
	return loan, nil
}