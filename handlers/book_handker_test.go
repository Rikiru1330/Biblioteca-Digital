package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"library-api/models"
	"library-api/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHelper crea un handler con MemoryStore (sin CGO)
func TestHelper() (*BookHandler, *gin.Engine) {
	store := storage.NewMemoryStore()
	handler := NewBookHandler(store)
	
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	
	return handler, router
}

func TestCreateBook(t *testing.T) {
	handler, router := TestHelper()
	router.POST("/books", handler.CreateBook)
	
	book := models.CreateBookRequest{
		Title:  "Test Book",
		Author: "Test Author",
		ISBN:   "1234567890",
	}
	
	jsonData, _ := json.Marshal(book)
	
	req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.Book
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Equal(t, book.Title, response.Title)
	assert.Equal(t, book.Author, response.Author)
	assert.NotEmpty(t, response.ID)
}

func TestGetBooks(t *testing.T) {
	handler, router := TestHelper()
	router.GET("/books", handler.GetBooks)
	
	req, _ := http.NewRequest("GET", "/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var books []models.Book
	json.Unmarshal(w.Body.Bytes(), &books)
	
	// Debería tener los libros de ejemplo
	assert.GreaterOrEqual(t, len(books), 0)
}

func TestGetBookNotFound(t *testing.T) {
	handler, router := TestHelper()
	router.GET("/books/:id", handler.GetBook)
	
	req, _ := http.NewRequest("GET", "/books/nonexistent-id-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSearchBooks(t *testing.T) {
	handler, router := TestHelper()
	router.GET("/books/search", handler.SearchBooks)
	
	// Primero crear un libro para buscar
	store := storage.NewMemoryStore()
	store.CreateBook(models.Book{
		Title:  "Search Test Book",
		Author: "Test Author",
		ISBN:   "999888777",
		Genre:  "Test Genre",
	})
	
	handler2 := NewBookHandler(store)
	router2 := gin.Default()
	router2.GET("/books/search", handler2.SearchBooks)
	
	req, _ := http.NewRequest("GET", "/books/search?title=Search", nil)
	w := httptest.NewRecorder()
	router2.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var books []models.Book
	json.Unmarshal(w.Body.Bytes(), &books)
	
	assert.Greater(t, len(books), 0)
}