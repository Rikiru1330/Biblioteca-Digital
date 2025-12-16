package storage

import (
	"library-api/models"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MemoryStore struct {
	books map[string]models.Book
	loans map[string]models.Loan
	users map[string]models.User
	mu    sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	// Hashear la contraseña del admin
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	return &MemoryStore{
		users: map[string]models.User{
			"1": {
				ID:        "1",
				Username:  "admin",
				Password:  string(hashedPassword), // ← CONTRASEÑA HASHED
				Role:      "admin",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		books: make(map[string]models.Book),
		loans: make(map[string]models.Loan),
	}
}

// ==============================================
// MÉTODOS PARA USUARIOS (CORREGIDOS PARA DEVOLVER PUNTEROS)
// ==============================================

// CreateUser - Crear un nuevo usuario (DEVUELVE PUNTERO)
func (s *MemoryStore) CreateUser(user models.User) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar si el usuario ya existe
	for _, u := range s.users {
		if u.Username == user.Username {
			return nil, ErrUserAlreadyExists
		}
	}

	// Generar ID si no tiene
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	s.users[user.ID] = user
	return &user, nil // ← CORREGIDO: devolver puntero
}

// GetUserByUsername - Obtener usuario por nombre de usuario (DEVUELVE PUNTERO)
func (s *MemoryStore) GetUserByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			// Crear una copia para devolver puntero a variable local
			userCopy := user
			return &userCopy, nil // ← CORREGIDO: devolver puntero
		}
	}

	return nil, ErrUserNotFound
}

// GetUserByID - Obtener usuario por ID (DEVUELVE PUNTERO)
func (s *MemoryStore) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	userCopy := user
	return &userCopy, nil // ← CORREGIDO: devolver puntero
}

// UpdateUser - Actualizar usuario (DEVUELVE PUNTERO)
func (s *MemoryStore) UpdateUser(id string, updatedUser models.User) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	updatedUser.ID = id
	updatedUser.CreatedAt = user.CreatedAt
	updatedUser.UpdatedAt = time.Now()

	s.users[id] = updatedUser

	updatedUserCopy := updatedUser
	return &updatedUserCopy, nil // ← CORREGIDO: devolver puntero
}

// DeleteUser - Eliminar usuario
func (s *MemoryStore) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return ErrUserNotFound
	}

	delete(s.users, id)
	return nil
}

// ==============================================
// MÉTODOS PARA LIBROS (CORREGIDOS PARA DEVOLVER PUNTEROS)
// ==============================================

// CreateBook - Crear un nuevo libro (DEVUELVE PUNTERO)
func (s *MemoryStore) CreateBook(book models.Book) (*models.Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	book.ID = uuid.New().String()
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	book.Available = true

	s.books[book.ID] = book
	return &book, nil // ← CORREGIDO: devolver puntero
}

// GetBooks - Obtener todos los libros
func (s *MemoryStore) GetBooks() ([]models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	books := make([]models.Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	return books, nil
}

// GetBookByID - Obtener libro por ID (DEVUELVE PUNTERO)
func (s *MemoryStore) GetBookByID(id string) (*models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, exists := s.books[id]
	if !exists {
		return nil, ErrBookNotFound
	}

	bookCopy := book
	return &bookCopy, nil // ← CORREGIDO: devolver puntero
}

// UpdateBook - Actualizar libro (DEVUELVE PUNTERO)
func (s *MemoryStore) UpdateBook(id string, updatedBook models.Book) (*models.Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	book, exists := s.books[id]
	if !exists {
		return nil, ErrBookNotFound
	}

	updatedBook.ID = id
	updatedBook.CreatedAt = book.CreatedAt
	updatedBook.UpdatedAt = time.Now()

	s.books[id] = updatedBook

	updatedBookCopy := updatedBook
	return &updatedBookCopy, nil // ← CORREGIDO: devolver puntero
}

// DeleteBook - Eliminar libro
func (s *MemoryStore) DeleteBook(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.books[id]; !exists {
		return ErrBookNotFound
	}

	delete(s.books, id)
	return nil
}

// SearchBooks - Buscar libros
func (s *MemoryStore) SearchBooks(title, author, genre string, available *bool) ([]models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []models.Book
	for _, book := range s.books {
		// Filtrar por texto
		matchesText := (title == "" || contains(book.Title, title)) &&
			(author == "" || contains(book.Author, author)) &&
			(genre == "" || contains(book.Genre, genre))

		// Filtrar por disponibilidad si se especifica
		matchesAvailability := true
		if available != nil {
			matchesAvailability = book.Available == *available
		}

		if matchesText && matchesAvailability {
			results = append(results, book)
		}
	}
	return results, nil
}

// UpdateBookAvailability - Actualizar disponibilidad de libro
func (s *MemoryStore) UpdateBookAvailability(bookID string, available bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	book, exists := s.books[bookID]
	if !exists {
		return ErrBookNotFound
	}

	book.Available = available
	book.UpdatedAt = time.Now()
	s.books[bookID] = book

	return nil
}

// ==============================================
// MÉTODOS PARA PRÉSTAMOS (CORREGIDOS PARA DEVOLVER PUNTEROS)
// ==============================================

// CreateLoan - Crear préstamo (DEVUELVE PUNTERO)
func (s *MemoryStore) CreateLoan(loan models.Loan) (*models.Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar que el libro existe y está disponible
	book, exists := s.books[loan.BookID]
	if !exists {
		return nil, ErrBookNotFound
	}
	if !book.Available {
		return nil, ErrBookNotAvailable
	}

	// Crear el préstamo
	loan.ID = uuid.New().String()
	loan.LoanDate = time.Now()
	loan.Returned = false

	s.loans[loan.ID] = loan

	// Marcar libro como no disponible
	book.Available = false
	book.UpdatedAt = time.Now()
	s.books[loan.BookID] = book

	return &loan, nil // ← CORREGIDO: devolver puntero
}

// ReturnBook - Devolver libro
func (s *MemoryStore) ReturnBook(loanID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	loan, exists := s.loans[loanID]
	if !exists {
		return ErrLoanNotFound
	}

	if loan.Returned {
		return nil // Ya estaba devuelto
	}

	// Marcar préstamo como devuelto
	now := time.Now()
	loan.Returned = true
	loan.ReturnDate = &now
	s.loans[loanID] = loan

	// Marcar libro como disponible
	if book, exists := s.books[loan.BookID]; exists {
		book.Available = true
		book.UpdatedAt = now
		s.books[loan.BookID] = book
	}

	return nil
}

// GetLoans - Obtener todos los préstamos
func (s *MemoryStore) GetLoans() ([]models.Loan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	loans := make([]models.Loan, 0, len(s.loans))
	for _, loan := range s.loans {
		loans = append(loans, loan)
	}
	return loans, nil
}

// GetActiveLoans - Obtener préstamos activos
func (s *MemoryStore) GetActiveLoans() ([]models.Loan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var activeLoans []models.Loan
	for _, loan := range s.loans {
		if !loan.Returned {
			activeLoans = append(activeLoans, loan)
		}
	}
	return activeLoans, nil
}

// GetLoanByID - Obtener préstamo por ID (DEVUELVE PUNTERO)
func (s *MemoryStore) GetLoanByID(id string) (*models.Loan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	loan, exists := s.loans[id]
	if !exists {
		return nil, ErrLoanNotFound
	}

	loanCopy := loan
	return &loanCopy, nil // ← CORREGIDO: devolver puntero
}

// GetLoansWithBooks - Obtener préstamos con información de libros
func (s *MemoryStore) GetLoansWithBooks() ([]models.LoanWithBook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var loansWithBooks []models.LoanWithBook

	for _, loan := range s.loans {
		// Buscar el libro correspondiente
		book, exists := s.books[loan.BookID]
		if !exists {
			// Si no encontramos el libro, crear uno vacío
			book = models.Book{
				ID:        loan.BookID,
				Title:     "Libro no encontrado",
				Available: false,
			}
		}

		loanWithBook := models.LoanWithBook{
			Loan: loan,
			Book: book,
		}

		loansWithBooks = append(loansWithBooks, loanWithBook)
	}

	return loansWithBooks, nil
}

// GetActiveLoansWithBooks - Obtener préstamos activos con información de libros
func (s *MemoryStore) GetActiveLoansWithBooks() ([]models.LoanWithBook, error) {
	allLoans, err := s.GetLoansWithBooks()
	if err != nil {
		return nil, err
	}

	var activeLoans []models.LoanWithBook
	for _, loan := range allLoans {
		if !loan.Returned {
			activeLoans = append(activeLoans, loan)
		}
	}

	return activeLoans, nil
}

// ==============================================
// FUNCIONES AUXILIARES
// ==============================================

// Función helper para búsqueda
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}

	s = strings.ToLower(s)
	substr = strings.ToLower(substr)

	return strings.Contains(s, substr)
}
