package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"library-api/models"
)

// ExternalBookService - Servicio para consumir APIs de libros externas
type ExternalBookService interface {
	SearchGoogleBooks(query string, maxResults int) ([]models.Book, error)
	SearchOpenLibrary(query string, limit int) ([]models.Book, error)
	GetGoogleBook(bookID string) (models.Book, error)
	GetOpenLibraryBook(bookID string) (models.Book, error)
	SearchGoogleBooksByISBN(isbn string, maxResults int) ([]models.Book, error)
	SearchOpenLibraryByISBN(isbn string, limit int) ([]models.Book, error)
}

// externalBookServiceImpl - Implementación concreta
type externalBookServiceImpl struct {
	googleAPIKey string
	client       *http.Client
}

// NewExternalBookService - Constructor
func NewExternalBookService(googleAPIKey string) ExternalBookService {
	return &externalBookServiceImpl{
		googleAPIKey: googleAPIKey,
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}
}

// ==============================================
// GOOGLE BOOKS API
// ==============================================

type googleBookItem struct {
	ID         string `json:"id"`
	VolumeInfo struct {
		Title               string   `json:"title"`
		Authors             []string `json:"authors"`
		Publisher           string   `json:"publisher"`
		PublishedDate       string   `json:"publishedDate"`
		Description         string   `json:"description"`
		IndustryIdentifiers []struct {
			Type       string `json:"type"`
			Identifier string `json:"identifier"`
		} `json:"industryIdentifiers"`
		PageCount  int      `json:"pageCount"`
		Categories []string `json:"categories"`
		ImageLinks struct {
			Thumbnail string `json:"thumbnail"`
		} `json:"imageLinks"`
		Language string `json:"language"`
	} `json:"volumeInfo"`
}

func (s *externalBookServiceImpl) SearchGoogleBooks(query string, maxResults int) ([]models.Book, error) {
	baseURL := "https://www.googleapis.com/books/v1/volumes"

	params := url.Values{}
	params.Add("q", query)
	params.Add("maxResults", strconv.Itoa(maxResults))
	params.Add("orderBy", "relevance")
	if s.googleAPIKey != "" {
		params.Add("key", s.googleAPIKey)
	}

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Library-API/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Google Books API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google Books API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Items []googleBookItem `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return s.convertGoogleBooks(result.Items), nil
}

func (s *externalBookServiceImpl) GetGoogleBook(bookID string) (models.Book, error) {
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes/%s", bookID)
	if s.googleAPIKey != "" {
		url += fmt.Sprintf("?key=%s", s.googleAPIKey)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.Book{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Library-API/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return models.Book{}, fmt.Errorf("error calling Google Books API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return models.Book{}, fmt.Errorf("book not found")
		}
		body, _ := io.ReadAll(resp.Body)
		return models.Book{}, fmt.Errorf("Google Books API error: %s - %s", resp.Status, string(body))
	}

	var item googleBookItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return models.Book{}, fmt.Errorf("error decoding response: %v", err)
	}

	return s.convertGoogleBook(item), nil
}

func (s *externalBookServiceImpl) SearchGoogleBooksByISBN(isbn string, maxResults int) ([]models.Book, error) {
	return s.SearchGoogleBooks(fmt.Sprintf("isbn:%s", isbn), maxResults)
}

// ==============================================
// OPEN LIBRARY API
// ==============================================

type openLibraryDoc struct {
	Key           string   `json:"key"`
	Title         string   `json:"title"`
	AuthorName    []string `json:"author_name"`
	PublishYear   []int    `json:"publish_year"`
	ISBN          []string `json:"isbn"`
	Subject       []string `json:"subject"`
	NumberOfPages int      `json:"number_of_pages"`
	Description   string   `json:"description"`
}

func (s *externalBookServiceImpl) SearchOpenLibrary(query string, limit int) ([]models.Book, error) {
	baseURL := "https://openlibrary.org/search.json"

	params := url.Values{}
	params.Add("q", query)
	params.Add("limit", strconv.Itoa(limit))
	params.Add("fields", "key,title,author_name,publish_year,isbn,subject,number_of_pages,description")

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Library-API/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Open Library API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open Library API error: %s", resp.Status)
	}

	var result struct {
		Docs []openLibraryDoc `json:"docs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return s.convertOpenLibraryBooks(result.Docs), nil
}

func (s *externalBookServiceImpl) GetOpenLibraryBook(bookID string) (models.Book, error) {
	// Para Open Library, usamos búsqueda por ID/título
	books, err := s.SearchOpenLibrary(bookID, 1)
	if err != nil {
		return models.Book{}, err
	}

	if len(books) == 0 {
		return models.Book{}, fmt.Errorf("book not found")
	}

	return books[0], nil
}

func (s *externalBookServiceImpl) SearchOpenLibraryByISBN(isbn string, limit int) ([]models.Book, error) {
	return s.SearchOpenLibrary(fmt.Sprintf("isbn:%s", isbn), limit)
}

// ==============================================
// CONVERSORES
// ==============================================

func (s *externalBookServiceImpl) convertGoogleBooks(items []googleBookItem) []models.Book {
	var books []models.Book

	for _, item := range items {
		book := s.convertGoogleBook(item)
		if book.Title != "" {
			books = append(books, book)
		}
	}

	return books
}

func (s *externalBookServiceImpl) convertGoogleBook(item googleBookItem) models.Book {
	// Extraer año de publicación
	publishedYear := 0
	if item.VolumeInfo.PublishedDate != "" && len(item.VolumeInfo.PublishedDate) >= 4 {
		if year, err := strconv.Atoi(item.VolumeInfo.PublishedDate[:4]); err == nil {
			publishedYear = year
		}
	}

	// Extraer ISBN
	isbn := ""
	for _, id := range item.VolumeInfo.IndustryIdentifiers {
		if id.Type == "ISBN_13" {
			isbn = id.Identifier
			break
		} else if id.Type == "ISBN_10" && isbn == "" {
			isbn = id.Identifier
		}
	}

	// Unir autores
	authors := ""
	if len(item.VolumeInfo.Authors) > 0 {
		authors = strings.Join(item.VolumeInfo.Authors, ", ")
	}

	// Unir categorías
	genre := ""
	if len(item.VolumeInfo.Categories) > 0 {
		genre = strings.Join(item.VolumeInfo.Categories, ", ")
	}

	// Limpiar descripción (puede tener HTML)
	description := item.VolumeInfo.Description
	if len(description) > 500 {
		description = description[:500] + "..."
	}

	return models.Book{
		ID:          item.ID,
		Title:       item.VolumeInfo.Title,
		Author:      authors,
		ISBN:        isbn,
		Published:   publishedYear,
		Genre:       genre,
		Description: description,
		Available:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (s *externalBookServiceImpl) convertOpenLibraryBooks(docs []openLibraryDoc) []models.Book {
	var books []models.Book

	for _, doc := range docs {
		book := s.convertOpenLibraryBook(doc)
		if book.Title != "" {
			books = append(books, book)
		}
	}

	return books
}

func (s *externalBookServiceImpl) convertOpenLibraryBook(doc openLibraryDoc) models.Book {
	// Tomar el primer año de publicación
	published := 0
	if len(doc.PublishYear) > 0 {
		published = doc.PublishYear[0]
	}

	// Buscar ISBN-13 primero, luego ISBN-10
	isbn := ""
	if len(doc.ISBN) > 0 {
		// Preferir ISBN de 13 dígitos
		for _, isbnStr := range doc.ISBN {
			if len(isbnStr) == 13 {
				isbn = isbnStr
				break
			}
		}
		// Si no hay ISBN-13, tomar el primero
		if isbn == "" {
			isbn = doc.ISBN[0]
		}
	}

	// Unir autores
	author := ""
	if len(doc.AuthorName) > 0 {
		author = strings.Join(doc.AuthorName, ", ")
	}

	// Unir géneros
	genre := ""
	if len(doc.Subject) > 0 {
		// Tomar solo algunos géneros únicos
		seen := make(map[string]bool)
		uniqueSubjects := []string{}
		for _, subject := range doc.Subject {
			if !seen[subject] {
				seen[subject] = true
				uniqueSubjects = append(uniqueSubjects, subject)
				if len(uniqueSubjects) >= 3 {
					break
				}
			}
		}
		genre = strings.Join(uniqueSubjects, ", ")
	}

	// Limpiar descripción
	description := doc.Description
	if description != "" {
		// Limitar longitud
		if len(description) > 400 {
			description = description[:400] + "..."
		}
	}

	return models.Book{
		ID:          doc.Key,
		Title:       doc.Title,
		Author:      author,
		ISBN:        isbn,
		Published:   published,
		Genre:       genre,
		Description: description,
		Available:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
