package storage

import (
	"database/sql"
	"fmt"
	"library-api/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sqlx.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	// Crear usuario admin por defecto si no existe
	store := &SQLiteStore{db: db}
	_, _ = store.ensureAdminUser()

	return store, nil
}

// ensureAdminUser - Crear usuario admin si no existe
func (s *SQLiteStore) ensureAdminUser() (*models.User, error) {
	existingUser, err := s.GetUserByUsername("admin")
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	adminUser := models.User{
		ID:       "1",
		Username: "admin",
		Password: "admin123", // En producción usar hash
		Role:     "admin",
	}

	return s.CreateUser(adminUser)
}

func createTables(db *sqlx.DB) error {
	// Tabla de usuarios (NUEVA)
	usersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        role TEXT DEFAULT 'user',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `

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
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);",
		"CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);",
		"CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);",
		"CREATE INDEX IF NOT EXISTS idx_books_available ON books(available);",
		"CREATE INDEX IF NOT EXISTS idx_loans_book_id ON loans(book_id);",
		"CREATE INDEX IF NOT EXISTS idx_loans_returned ON loans(returned);",
	}

	// Ejecutar creación de tablas
	tables := []string{usersTable, booksTable, loansTable}
	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("error creating table: %w", err)
		}
	}

	// Crear índices
	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
	}

	return nil
}

// ==============================================
// MÉTODOS PARA USUARIOS (CORREGIDOS PARA DEVOLVER PUNTEROS)
// ==============================================

// CreateUser - Crear un nuevo usuario (DEVUELVE PUNTERO)
func (s *SQLiteStore) CreateUser(user models.User) (*models.User, error) {
	// Verificar si el usuario ya existe
	existingUser, _ := s.GetUserByUsername(user.Username)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Generar ID si no tiene
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, username, password, role, created_at, updated_at) 
              VALUES (:id, :username, :password, :role, :created_at, :updated_at)`

	_, err := s.db.NamedExec(query, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil // ← CORREGIDO: devolver puntero
}

// GetUserByUsername - Obtener usuario por nombre de usuario (DEVUELVE PUNTERO)
func (s *SQLiteStore) GetUserByUsername(username string) (*models.User, error) {
	var user models.User

	query := `SELECT * FROM users WHERE username = ? LIMIT 1`
	err := s.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil // ← CORREGIDO: devolver puntero
}

// GetUserByID - Obtener usuario por ID (DEVUELVE PUNTERO)
func (s *SQLiteStore) GetUserByID(id string) (*models.User, error) {
	var user models.User

	query := `SELECT * FROM users WHERE id = ? LIMIT 1`
	err := s.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil // ← CORREGIDO: devolver puntero
}

// UpdateUser - Actualizar usuario (DEVUELVE PUNTERO)
func (s *SQLiteStore) UpdateUser(id string, user models.User) (*models.User, error) {
	user.UpdatedAt = time.Now()

	query := `UPDATE users SET 
        username = :username, 
        password = :password, 
        role = :role, 
        updated_at = :updated_at 
        WHERE id = :id`

	user.ID = id
	_, err := s.db.NamedExec(query, user)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return s.GetUserByID(id)
}

// DeleteUser - Eliminar usuario
func (s *SQLiteStore) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ==============================================
// MÉTODOS PARA LIBROS (CORREGIDOS PARA DEVOLVER PUNTEROS)
// ==============================================

// CreateBook implementación (DEVUELVE PUNTERO)
func (s *SQLiteStore) CreateBook(book models.Book) (*models.Book, error) {
	book.ID = uuid.New().String()
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	book.Available = true

	query := `INSERT INTO books (id, title, author, isbn, published, genre, description, available, created_at, updated_at) 
              VALUES (:id, :title, :author, :isbn, :published, :genre, :description, :available, :created_at, :updated_at)`

	_, err := s.db.NamedExec(query, book)
	if err != nil {
		return nil, fmt.Errorf("error creating book: %w", err)
	}

	return &book, nil // ← CORREGIDO: devolver puntero
}

// GetBooks implementación
func (s *SQLiteStore) GetBooks() ([]models.Book, error) {
	var books []models.Book
	query := `SELECT 
        id, 
        title, 
        author, 
        isbn, 
        published, 
        genre, 
        description, 
        available, 
        created_at, 
        updated_at 
        FROM books ORDER BY title`

	err := s.db.Select(&books, query)
	if err != nil {
		return nil, fmt.Errorf("error getting books: %w", err)
	}

	return books, nil
}

// GetBookByID implementación (DEVUELVE PUNTERO)
func (s *SQLiteStore) GetBookByID(id string) (*models.Book, error) {
	var book models.Book
	query := `SELECT * FROM books WHERE id = ?`

	err := s.db.Get(&book, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("error getting book: %w", err)
	}

	return &book, nil // ← CORREGIDO: devolver puntero
}

// UpdateBook implementación (DEVUELVE PUNTERO)
func (s *SQLiteStore) UpdateBook(id string, updatedBook models.Book) (*models.Book, error) {
	// Obtener libro existente
	existingBook, err := s.GetBookByID(id)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("error updating book: %w", err)
	}

	return &updatedBook, nil // ← CORREGIDO: devolver puntero
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

// UpdateBookAvailability - Actualizar disponibilidad de libro
func (s *SQLiteStore) UpdateBookAvailability(bookID string, available bool) error {
	query := `UPDATE books SET available = ?, updated_at = ? WHERE id = ?`

	_, err := s.db.Exec(query, available, time.Now(), bookID)
	if err != nil {
		return fmt.Errorf("error updating book availability: %w", err)
	}

	return nil
}

// ==============================================
// MÉTODOS PARA PRÉSTAMOS (CORREGIDOS PARA DEVOLVER PUNTEROS)
// ==============================================

// CreateLoan implementación (DEVUELVE PUNTERO)
func (s *SQLiteStore) CreateLoan(loan models.Loan) (*models.Loan, error) {
	// Verificar que el libro existe y está disponible
	var available bool
	checkQuery := `SELECT available FROM books WHERE id = ?`
	err := s.db.Get(&available, checkQuery, loan.BookID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("error checking book availability: %w", err)
	}

	if !available {
		return nil, ErrBookNotAvailable
	}

	// Crear el préstamo
	loan.ID = uuid.New().String()
	loan.LoanDate = time.Now()
	loan.Returned = false

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Insertar préstamo - ESPECIFICAR COLUMNAS EXPLÍCITAMENTE
	loanQuery := `INSERT INTO loans (id, book_id, user, loan_date, returned) 
                  VALUES (:id, :book_id, :user, :loan_date, :returned)`

	// Usar un mapa para asegurar el mapeo correcto
	loanMap := map[string]interface{}{
		"id":        loan.ID,
		"book_id":   loan.BookID, // Asegurar que se mapea a book_id
		"user":      loan.User,
		"loan_date": loan.LoanDate,
		"returned":  loan.Returned,
	}

	_, err = tx.NamedExec(loanQuery, loanMap)
	if err != nil {
		return nil, fmt.Errorf("error creating loan: %w", err)
	}

	// Actualizar libro como no disponible
	updateBookQuery := `UPDATE books SET available = FALSE, updated_at = ? WHERE id = ?`
	_, err = tx.Exec(updateBookQuery, time.Now(), loan.BookID)
	if err != nil {
		return nil, fmt.Errorf("error updating book status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &loan, nil // ← CORREGIDO: devolver puntero
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

// GetLoanByID implementación (DEVUELVE PUNTERO)
func (s *SQLiteStore) GetLoanByID(id string) (*models.Loan, error) {
	var loan models.Loan
	query := `SELECT * FROM loans WHERE id = ?`

	err := s.db.Get(&loan, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLoanNotFound
		}
		return nil, fmt.Errorf("error getting loan: %w", err)
	}

	return &loan, nil // ← CORREGIDO: devolver puntero
}

// GetLoansWithBooks - Obtener préstamos con información de libros
func (s *SQLiteStore) GetLoansWithBooks() ([]models.LoanWithBook, error) {
	query := `
    SELECT 
        l.*,
        COALESCE(b.id, 'DELETED') as "book.id",
        COALESCE(b.title, 'Libro eliminado') as "book.title",
        COALESCE(b.author, 'N/A') as "book.author",
        COALESCE(b.isbn, 'N/A') as "book.isbn",
        COALESCE(b.published, 0) as "book.published",
        COALESCE(b.genre, 'N/A') as "book.genre",
        COALESCE(b.description, 'Este libro ha sido eliminado') as "book.description",
        COALESCE(b.available, FALSE) as "book.available",
        COALESCE(b.created_at, l.loan_date) as "book.created_at",
        COALESCE(b.updated_at, l.loan_date) as "book.updated_at"
    FROM loans l
    LEFT JOIN books b ON l.book_id = b.id
    ORDER BY l.loan_date DESC
    `

	var loansWithBooks []models.LoanWithBook
	err := s.db.Select(&loansWithBooks, query)
	if err != nil {
		// Si es "no rows", retornar slice vacío, no error
		if err == sql.ErrNoRows {
			return []models.LoanWithBook{}, nil
		}
		return nil, fmt.Errorf("error getting loans with books: %w", err)
	}

	// Asegurar que no sea nil
	if loansWithBooks == nil {
		return []models.LoanWithBook{}, nil
	}

	return loansWithBooks, nil
}

// GetActiveLoansWithBooks - Obtener préstamos activos con información de libros
func (s *SQLiteStore) GetActiveLoansWithBooks() ([]models.LoanWithBook, error) {
	query := `
    SELECT 
        l.*,
        COALESCE(b.id, 'DELETED') as "book.id",
        COALESCE(b.title, 'Libro eliminado') as "book.title",
        COALESCE(b.author, 'N/A') as "book.author",
        COALESCE(b.isbn, 'N/A') as "book.isbn",
        COALESCE(b.published, 0) as "book.published",
        COALESCE(b.genre, 'N/A') as "book.genre",
        COALESCE(b.description, 'Este libro ha sido eliminado') as "book.description",
        COALESCE(b.available, FALSE) as "book.available",
        COALESCE(b.created_at, l.loan_date) as "book.created_at",
        COALESCE(b.updated_at, l.loan_date) as "book.updated_at"
    FROM loans l
    LEFT JOIN books b ON l.book_id = b.id
    WHERE l.returned = FALSE
    ORDER BY l.loan_date DESC
    `

	var loansWithBooks []models.LoanWithBook
	err := s.db.Select(&loansWithBooks, query)
	if err != nil {
		return nil, fmt.Errorf("error getting active loans with books: %w", err)
	}

	return loansWithBooks, nil
}
