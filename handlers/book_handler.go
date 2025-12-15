package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"library-api/models"
	"library-api/storage"

	"github.com/gin-gonic/gin"
)

// Interface para servicio externo
type ExternalBookService interface {
	SearchGoogleBooks(query string, maxResults int) ([]models.Book, error)
	SearchOpenLibrary(query string, limit int) ([]models.Book, error)
	GetGoogleBook(bookID string) (models.Book, error)
	GetOpenLibraryBook(bookID string) (models.Book, error)
	SearchGoogleBooksByISBN(isbn string, maxResults int) ([]models.Book, error)
	SearchOpenLibraryByISBN(isbn string, limit int) ([]models.Book, error)
}

type BookHandler struct {
	store           storage.Store
	externalService ExternalBookService
}

func NewBookHandler(store storage.Store, externalService ExternalBookService) *BookHandler {
	return &BookHandler{
		store:           store,
		externalService: externalService,
	}
}

// ==============================================
// MÉTODOS EXISTENTES (modificados solo el constructor)
// ==============================================

// CreateBook - Crear un nuevo libro
func (h *BookHandler) CreateBook(c *gin.Context) {
	var req models.CreateBookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book := models.Book{
		Title:       req.Title,
		Author:      req.Author,
		ISBN:        req.ISBN,
		Published:   req.Published,
		Genre:       req.Genre,
		Description: req.Description,
	}

	createdBook, err := h.store.CreateBook(book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating book: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdBook)
}

// GetBooks - Obtener todos los libros
func (h *BookHandler) GetBooks(c *gin.Context) {
	books, err := h.store.GetBooks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting books: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, books)
}

// GetBook - Obtener un libro por ID
func (h *BookHandler) GetBook(c *gin.Context) {
	id := c.Param("id")

	book, err := h.store.GetBookByID(id)
	if err != nil {
		if err == storage.ErrBookNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting book: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, book)
}

// UpdateBook - Actualizar un libro
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingBook, err := h.store.GetBookByID(id)
	if err != nil {
		if err == storage.ErrBookNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting book: " + err.Error()})
		}
		return
	}

	if req.Title != "" {
		existingBook.Title = req.Title
	}
	if req.Author != "" {
		existingBook.Author = req.Author
	}
	if req.ISBN != "" {
		existingBook.ISBN = req.ISBN
	}
	if req.Published != 0 {
		existingBook.Published = req.Published
	}
	if req.Genre != "" {
		existingBook.Genre = req.Genre
	}
	if req.Description != "" {
		existingBook.Description = req.Description
	}
	if req.Available != nil {
		existingBook.Available = *req.Available
	}

	updatedBook, err := h.store.UpdateBook(id, existingBook)
	if err != nil {
		if err == storage.ErrBookNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating book: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedBook)
}

// DeleteBook - Eliminar un libro
func (h *BookHandler) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteBook(id); err != nil {
		if err == storage.ErrBookNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting book: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

// SearchBooks - Buscar libros en nuestra base
func (h *BookHandler) SearchBooks(c *gin.Context) {
	title := c.Query("title")
	author := c.Query("author")
	genre := c.Query("genre")
	availableStr := c.Query("available")

	var available *bool
	if availableStr != "" {
		avail, err := strconv.ParseBool(availableStr)
		if err == nil {
			available = &avail
		}
	}

	books, err := h.store.SearchBooks(title, author, genre, available)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching books: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, books)
}

// BorrowBook - Prestar un libro
func (h *BookHandler) BorrowBook(c *gin.Context) {
	var req models.LoanRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan := models.Loan{
		BookID: req.BookID,
		User:   req.User,
	}

	createdLoan, err := h.store.CreateLoan(loan)
	if err != nil {
		if err == storage.ErrBookNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else if err == storage.ErrBookNotAvailable {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Book is not available"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdLoan)
}

// ReturnBook - Devolver un libro
func (h *BookHandler) ReturnBook(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.ReturnBook(id); err != nil {
		if err == storage.ErrLoanNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book returned successfully"})
}

// GetLoans - Obtener todos los préstamos
func (h *BookHandler) GetLoans(c *gin.Context) {
	status := c.Query("status")
	var loans []models.Loan
	var err error

	if strings.ToLower(status) == "active" {
		loans, err = h.store.GetActiveLoans()
	} else {
		loans, err = h.store.GetLoans()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting loans: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// HealthCheck - Verificar estado del API
func (h *BookHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Library API is running",
	})
}

// ==============================================
// NUEVOS MÉTODOS PARA APIS EXTERNAS
// ==============================================

// SearchExternalBooks - Buscar libros en APIs externas
func (h *BookHandler) SearchExternalBooks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	source := c.DefaultQuery("source", "openlibrary")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	var books []models.Book

	switch source {
	case "google":
		books, err = h.externalService.SearchGoogleBooks(query, limit)
	case "openlibrary":
		books, err = h.externalService.SearchOpenLibrary(query, limit)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source. Use 'google' or 'openlibrary'"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source":  source,
		"query":   query,
		"results": books,
	})
}

// ImportBookFromExternal - Importar un libro desde API externa
func (h *BookHandler) ImportBookFromExternal(c *gin.Context) {
	source := c.Query("source")
	externalID := c.Query("id")

	if source == "" || externalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Parameters 'source' and 'id' are required",
			"example": "/api/external/import?source=openlibrary&id=OL1234567M",
		})
		return
	}

	var book models.Book
	var err error

	switch source {
	case "google":
		book, err = h.externalService.GetGoogleBook(externalID)
	case "openlibrary":
		// Open Library usa búsqueda para obtener detalles
		books, searchErr := h.externalService.SearchOpenLibrary(externalID, 1)
		if searchErr != nil || len(books) == 0 {
			err = fmt.Errorf("book not found in Open Library")
		} else {
			book = books[0]
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source. Use 'google' or 'openlibrary'"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verificar si ya existe en nuestra base por ISBN
	if book.ISBN != "" {
		existingBooks, _ := h.store.SearchBooks("", "", "", nil)
		for _, existing := range existingBooks {
			if existing.ISBN == book.ISBN {
				c.JSON(http.StatusConflict, gin.H{
					"error":         "Book already exists in database",
					"existing_book": existing,
				})
				return
			}
		}
	}

	// Guardar en nuestra base de datos
	createdBook, err := h.store.CreateBook(book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving book: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Book imported successfully",
		"book":    createdBook,
		"source":  source,
	})
}

// BulkImportBooks - Importar múltiples libros desde búsqueda
func (h *BookHandler) BulkImportBooks(c *gin.Context) {
	var req struct {
		Query  string `json:"query" binding:"required"`
		Source string `json:"source" binding:"required"`
		Limit  int    `json:"limit"`
		Filter string `json:"filter"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 || req.Limit > 20 {
		req.Limit = 5
	}

	var externalBooks []models.Book
	var err error

	switch req.Source {
	case "google":
		externalBooks, err = h.externalService.SearchGoogleBooks(req.Query, req.Limit)
	case "openlibrary":
		externalBooks, err = h.externalService.SearchOpenLibrary(req.Query, req.Limit)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Aplicar filtro si se especificó
	if req.Filter != "" {
		filteredBooks := make([]models.Book, 0)
		for _, book := range externalBooks {
			if matchesFilter(book, req.Filter) {
				filteredBooks = append(filteredBooks, book)
			}
		}
		externalBooks = filteredBooks
	}

	// Importar solo los que no existen
	imported := make([]models.Book, 0)
	failed := make([]string, 0)

	for _, book := range externalBooks {
		// Verificar si ya existe por ISBN
		exists := false
		if book.ISBN != "" {
			existingBooks, _ := h.store.SearchBooks("", "", "", nil)
			for _, existing := range existingBooks {
				if existing.ISBN == book.ISBN {
					exists = true
					break
				}
			}
		}

		if !exists {
			createdBook, err := h.store.CreateBook(book)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s: %v", book.Title, err))
			} else {
				imported = append(imported, createdBook)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"imported":       len(imported),
		"already_exists": len(externalBooks) - len(imported) - len(failed),
		"failed":         len(failed),
		"imported_books": imported,
		"failed_books":   failed,
	})
}

// GetBookDetails - Obtener detalles extendidos de un libro (combinando fuentes)
func (h *BookHandler) GetBookDetails(c *gin.Context) {
	id := c.Param("id")

	// Primero buscar en nuestra base de datos
	book, err := h.store.GetBookByID(id)
	if err != nil {
		// Si no está en nuestra base, buscar en APIs externas
		source := c.Query("source")
		if source != "" {
			switch source {
			case "google":
				book, err = h.externalService.GetGoogleBook(id)
			case "openlibrary":
				books, searchErr := h.externalService.SearchOpenLibrary(id, 1)
				if searchErr == nil && len(books) > 0 {
					book = books[0]
					err = nil
				} else {
					err = fmt.Errorf("book not found")
				}
			}

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Book not found in any source"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"source":     source,
				"in_local":   false,
				"book":       book,
				"can_import": true,
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Si está en nuestra base, buscar información adicional en APIs externas
	enrichSource := c.Query("enrich")
	if enrichSource != "" && book.ISBN != "" {
		var enrichedBooks []models.Book
		switch enrichSource {
		case "google":
			enrichedBooks, _ = h.externalService.SearchGoogleBooksByISBN(book.ISBN, 1)
		case "openlibrary":
			enrichedBooks, _ = h.externalService.SearchOpenLibraryByISBN(book.ISBN, 1)
		}

		// Combinar información si encontramos
		if len(enrichedBooks) > 0 {
			enrichedBook := enrichedBooks[0]

			if enrichedBook.Title != "" && book.Title == "" {
				book.Title = enrichedBook.Title
			}
			if enrichedBook.Author != "" && book.Author == "" {
				book.Author = enrichedBook.Author
			}
			if enrichedBook.Description != "" && book.Description == "" {
				book.Description = enrichedBook.Description
			}
			if enrichedBook.Published > 0 && book.Published == 0 {
				book.Published = enrichedBook.Published
			}
			if enrichedBook.Genre != "" && book.Genre == "" {
				book.Genre = enrichedBook.Genre
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"source":   "local",
		"in_local": true,
		"book":     book,
	})
}

// ==============================================
// MÉTODOS AUXILIARES
// ==============================================

func matchesFilter(book models.Book, filter string) bool {
	switch filter {
	case "title":
		return book.Title != ""
	case "author":
		return book.Author != ""
	case "year":
		return book.Published > 0
	case "isbn":
		return book.ISBN != ""
	case "complete":
		return book.Title != "" && book.Author != "" && book.ISBN != ""
	default:
		return true
	}
}
