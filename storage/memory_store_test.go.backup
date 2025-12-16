package storage

import (
	"testing"
	"library-api/models"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStore_CreateAndGetBook(t *testing.T) {
	store := NewMemoryStore()
	
	book := models.Book{
		Title:  "Test Book",
		Author: "Test Author",
		ISBN:   "1234567890",
	}
	
	created, err := store.CreateBook(book)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, book.Title, created.Title)
	assert.True(t, created.Available)
	
	retrieved, err := store.GetBookByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
}

func TestMemoryStore_UpdateBook(t *testing.T) {
	store := NewMemoryStore()
	
	book := models.Book{
		Title:  "Original Title",
		Author: "Original Author",
		ISBN:   "1111111111",
	}
	
	created, _ := store.CreateBook(book)
	
	updatedBook := created
	updatedBook.Title = "Updated Title"
	updatedBook.Author = "Updated Author"
	
	result, err := store.UpdateBook(created.ID, updatedBook)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", result.Title)
	assert.Equal(t, "Updated Author", result.Author)
	assert.Equal(t, created.ID, result.ID)
}

func TestMemoryStore_DeleteBook(t *testing.T) {
	store := NewMemoryStore()
	
	book := models.Book{
		Title:  "To Delete",
		Author: "Author",
		ISBN:   "2222222222",
	}
	
	created, _ := store.CreateBook(book)
	
	err := store.DeleteBook(created.ID)
	assert.NoError(t, err)
	
	_, err = store.GetBookByID(created.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrBookNotFound, err)
}

func TestMemoryStore_CreateAndReturnLoan(t *testing.T) {
	store := NewMemoryStore()
	
	// Crear libro
	book, _ := store.CreateBook(models.Book{
		Title:  "Loan Test",
		Author: "Author",
		ISBN:   "9999999999",
	})
	
	// Verificar que está disponible
	assert.True(t, book.Available)
	
	// Crear préstamo
	loan, err := store.CreateLoan(models.Loan{
		BookID: book.ID,
		User:   "Test User",
	})
	
	assert.NoError(t, err)
	assert.NotEmpty(t, loan.ID)
	assert.False(t, loan.Returned)
	
	// Devolver libro
	err = store.ReturnBook(loan.ID)
	assert.NoError(t, err)
	
	// Verificar que el préstamo está marcado como devuelto
	updatedLoan, _ := store.GetLoanByID(loan.ID)
	assert.True(t, updatedLoan.Returned)
}