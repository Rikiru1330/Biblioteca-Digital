package main

import (
	"library-api/handlers"
	"library-api/middleware"
	"library-api/models"
	"library-api/services"
	"library-api/storage"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No se cargó el archivo .env, usando variables del sistema")
	}

	_ = time.Now()

	// Configurar desde variables de entorno
	port := getEnv("PORT", "8080")
	storageType := getEnv("STORAGE_TYPE", "sqlite") // Cambiado a sqlite por defecto
	ginMode := getEnv("GIN_MODE", "release")
	googleAPIKey := getEnv("GOOGLE_BOOKS_API_KEY", "")

	// Configurar Gin
	gin.SetMode(ginMode)

	// Crear store según configuración
	var store storage.Store
	if storageType == "sqlite" {
		dbPath := getEnv("DB_PATH", "./data/library.db")
		sqliteStore, err := storage.NewSQLiteStore(dbPath)
		if err != nil {
			log.Printf("⚠️  No se pudo inicializar SQLite: %v", err)
			log.Println("✅ Usando MemoryStore como fallback")
			store = storage.NewMemoryStore()
		} else {
			store = sqliteStore
			log.Println("✅ Usando SQLiteStore:", dbPath)
		}
	} else {
		store = storage.NewMemoryStore()
		log.Println("✅ Usando MemoryStore")
	}

	// Crear servicio externo de libros
	externalService := services.NewExternalBookService(googleAPIKey)
	if googleAPIKey == "" {
		log.Println("⚠️  GOOGLE_BOOKS_API_KEY no configurada, usando solo Open Library (gratuito)")
	} else {
		log.Println("✅ Google Books API configurada")
	}

	// Inicializar handlers CON el servicio externo
	bookHandler := handlers.NewBookHandler(store, externalService)
	authHandler := handlers.NewAuthHandler(store)

	// Agregar datos de ejemplo solo si no hay datos
	if err := addSampleData(store); err != nil {
		log.Println("⚠️ Warning:", err)
	}

	// Crear router
	router := gin.Default()

	// Middleware para headers UTF-8
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// Middleware estándar
	if ginMode != "release" {
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())

	// Configurar CORS
	router.Use(corsMiddleware())

	// ==================== RUTAS PÚBLICAS ====================
	setupRoutes(router, bookHandler, authHandler)

	// ==================== INICIAR SERVIDOR ====================
	fullPort := ":" + port
	log.Println("🚀 Server starting on port", port)
	log.Println("📦 Storage:", storageType)
	log.Println("🔐 Default: admin / admin123")
	log.Println("🔍 External APIs: Open Library" + func() string {
		if googleAPIKey != "" {
			return " + Google Books"
		}
		return ""
	}())
	log.Println("🌐 http://localhost:" + port)

	if err := router.Run(fullPort); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

// ==================== FUNCIONES AUXILIARES ====================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupRoutes(router *gin.Engine, bookHandler *handlers.BookHandler, authHandler *handlers.AuthHandler) {
	// Ruta raíz - Documentación de la API
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "📚 Library Digital API",
			"version":  "1.1.0",
			"features": "CRUD de libros + APIs externas (Open Library, Google Books)",
			"endpoints": map[string]string{
				"home":                   "GET /",
				"health":                 "GET /health",
				"auth_register":          "POST /register",
				"auth_login":             "POST /login",
				"books_list":             "GET /books",
				"book_detail":            "GET /books/:id",
				"book_search":            "GET /books/search?title=...&author=...",
				"external_search":        "GET /api/external/search?q=harry+potter&source=openlibrary",
				"external_import":        "GET /api/external/import?source=openlibrary&id=OL1234567M",
				"book_details":           "GET /api/books/:id/details?enrich=google",
				"protected_books_create": "POST /books (requiere auth)",
				"protected_bulk_import":  "POST /api/external/import/bulk (requiere auth)",
			},
		})
	})

	// Autenticación
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/api/register", authHandler.Register)

	// Salud del sistema
	router.GET("/health", bookHandler.HealthCheck)

	// ==================== RUTAS PÚBLICAS DE LIBROS ====================
	// Libros en nuestra base de datos
	router.GET("/books", bookHandler.GetBooks)
	router.GET("/books/search", bookHandler.SearchBooks)
	router.GET("/books/:id", bookHandler.GetBook)

	// ==================== NUEVAS RUTAS PARA APIS EXTERNAS ====================
	// Buscar en APIs externas (público)
	router.GET("/api/external/search", bookHandler.SearchExternalBooks)

	// Importar un libro desde API externa (público - pero solo para ver, el guardado requiere auth)
	router.GET("/api/external/import", bookHandler.ImportBookFromExternal)

	// Obtener detalles combinados (local + externo)
	router.GET("/api/books/:id/details", bookHandler.GetBookDetails)

	// ==================== RUTAS PROTEGIDAS ====================
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// CRUD de libros en nuestra base
		protected.POST("/books", bookHandler.CreateBook)
		protected.PUT("/books/:id", bookHandler.UpdateBook)
		protected.DELETE("/books/:id", bookHandler.DeleteBook)

		// Importación masiva desde APIs externas (protegida)
		protected.POST("/api/external/import/bulk", bookHandler.BulkImportBooks)

		// Sistema de préstamos
		protected.POST("/books/:id/borrow", bookHandler.BorrowBook)
		protected.POST("/loans/:id/return", bookHandler.ReturnBook)
		protected.GET("/loans", bookHandler.GetLoans)

		// Información del usuario autenticado
		protected.GET("/me", authHandler.Me)
	}

	// Ruta de documentación Swagger/OpenAPI (si la agregas después)
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"openapi": "3.0.0",
			"info": gin.H{
				"title":       "Library Digital API",
				"version":     "1.1.0",
				"description": "API para gestión de biblioteca digital con integración de APIs externas",
			},
			"externalDocs": gin.H{
				"description": "Open Library API",
				"url":         "https://openlibrary.org/developers/api",
			},
		})
	})
}

func addSampleData(store storage.Store) error {
	// Verificar si ya hay libros
	books, err := store.GetBooks()
	if err != nil {
		return err
	}

	// Solo agregar datos si no hay libros
	if len(books) == 0 {
		sampleBooks := []models.Book{
			{
				Title:       "Cien años de soledad",
				Author:      "Gabriel García Márquez",
				ISBN:        "978-0307474728",
				Published:   1967,
				Genre:       "Realismo mágico, Novela",
				Description: "Crónica de la familia Buendía en el pueblo ficticio de Macondo.",
				Available:   true,
			},
			{
				Title:       "1984",
				Author:      "George Orwell",
				ISBN:        "978-0451524935",
				Published:   1949,
				Genre:       "Distopía, Ciencia ficción política",
				Description: "Novela sobre vigilancia y control totalitario en un futuro distópico.",
				Available:   true,
			},
			{
				Title:       "Don Quijote de la Mancha",
				Author:      "Miguel de Cervantes",
				ISBN:        "978-8424113296",
				Published:   1605,
				Genre:       "Novela, Aventura, Sátira",
				Description: "Las aventuras de un hidalgo que enloquece leyendo libros de caballerías.",
				Available:   true,
			},
		}

		count := 0
		for _, book := range sampleBooks {
			if _, err := store.CreateBook(book); err == nil {
				count++
			} else {
				log.Printf("⚠️  Error creando libro de ejemplo: %v", err)
			}
		}

		if count > 0 {
			log.Printf("📚 Se agregaron %d libros de ejemplo", count)
		}

		// Crear usuario admin por defecto si no existe
		adminUser := models.User{
			Username: "admin",
			Password: "admin123", // En un caso real, esto estaría hasheado
			Role:     "admin",
		}

		if _, err := store.CreateUser(adminUser); err != nil {
			log.Printf("⚠️  Error creando usuario admin: %v", err)
		} else {
			log.Println("👤 Usuario admin creado (admin / admin123)")
		}
	}

	return nil
}

// Función auxiliar para crear respuesta estandarizada
func apiResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"status":    status >= 200 && status < 300,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
