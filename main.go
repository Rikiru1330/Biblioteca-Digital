package main

import (
	"log"
	"os"
	"time"
	"library-api/handlers"
	"library-api/middleware"
	"library-api/models"
	"library-api/storage"
	"github.com/gin-gonic/gin"
)

func main() {

	_ = time.Now()
	// Configurar desde variables de entorno
	port := getEnv("PORT", "8080")
	storageType := getEnv("STORAGE_TYPE", "memory") // memory o sqlite
	ginMode := getEnv("GIN_MODE", "release")
	
	// Configurar Gin
	gin.SetMode(ginMode)
	
	// Crear store según configuración
	var store storage.Store
	if storageType == "sqlite" {
		// Usar SQLite si está configurado y CGO está disponible
		dbPath := getEnv("DB_PATH", "data/library.db")
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
	
	// Inicializar handlers
	bookHandler := handlers.NewBookHandler(store)
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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func setupRoutes(router *gin.Engine, bookHandler *handlers.BookHandler, authHandler *handlers.AuthHandler) {
	// Ruta raíz
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "📚 Library API",
			"version":   "1.0.0",
			"endpoints": []string{"/books", "/login", "/register", "/health", "/docs"},
		})
	})
	
	// Autenticación
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	
	// Salud
	router.GET("/health", bookHandler.HealthCheck)
	
	// Libros públicos
	router.GET("/books", bookHandler.GetBooks)
	router.GET("/books/search", bookHandler.SearchBooks)
	router.GET("/books/:id", bookHandler.GetBook)
	
	// Rutas protegidas
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/books", bookHandler.CreateBook)
		protected.PUT("/books/:id", bookHandler.UpdateBook)
		protected.DELETE("/books/:id", bookHandler.DeleteBook)
		protected.POST("/books/:id/borrow", bookHandler.BorrowBook)
		
		protected.GET("/loans", bookHandler.GetLoans)
		protected.POST("/loans/:id/return", bookHandler.ReturnBook)
		
		protected.GET("/me", authHandler.Me)
	}
}

func addSampleData(store storage.Store) error {
	books := []models.Book{
		{
			Title:       "Cien años de soledad",
			Author:      "Gabriel García Márquez",
			ISBN:        "978-0307474728",
			Published:   1967,
			Genre:       "Realismo mágico",
			Description: "Crónica de la familia Buendía en Macondo",
		},
		{
			Title:       "1984",
			Author:      "George Orwell",
			ISBN:        "978-0451524935",
			Published:   1949,
			Genre:       "Distopía",
			Description: "Novela sobre vigilancia y control totalitario",
		},
	}
	
	count := 0
	for _, book := range books {
		if _, err := store.CreateBook(book); err == nil {
			count++
		}
	}
	
	if count > 0 {
		log.Printf("📚 Se agregaron %d libros de ejemplo", count)
	}
	
	return nil
}