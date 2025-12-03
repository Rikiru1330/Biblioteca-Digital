package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"library-api/models"
	"library-api/storage"
	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	store storage.Store  // Cambia a la interfaz, no al tipo concreto
}

func NewBookHandler(store storage.Store) *BookHandler {  // Cambia el parámetro
	return &BookHandler{store: store}
}

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
	
	createdBook, err := h.store.CreateBook(book)  // Ahora devuelve error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating book: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, createdBook)
}

// GetBooks - Obtener todos los libros
func (h *BookHandler) GetBooks(c *gin.Context) {
	books, err := h.store.GetBooks()  // Ahora devuelve error
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
	
	// Obtener libro existente
	existingBook, err := h.store.GetBookByID(id)
	if err != nil {
		if err == storage.ErrBookNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting book: " + err.Error()})
		}
		return
	}
	
	// Actualizar solo los campos proporcionados
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

// SearchBooks - Buscar libros
func (h *BookHandler) SearchBooks(c *gin.Context) {
	title := c.Query("title")
	author := c.Query("author")
	genre := c.Query("genre")
	availableStr := c.Query("available")
	
	// Convertir available string a *bool
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
		"status": "healthy",
		"message": "Library API is running",
	})
}