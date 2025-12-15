package handlers

import (
	"library-api/auth"
	"library-api/models"
	"library-api/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	store storage.Store
}

func NewAuthHandler(store storage.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

// Register - Registrar nuevo usuario
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing password"})
		return
	}

	// Crear usuario (en memoria por ahora, luego agregaremos al store)
	user := models.User{
		Username: req.Username,
		//Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user", // Por defecto
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Guardar usuario en la base de datos
	// Por ahora solo generamos el token

	token, err := auth.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// No devolvemos la contraseña
	user.Password = ""

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// Login - Iniciar sesión
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Buscar usuario en la base de datos
	// Por ahora usamos un usuario hardcodeado para pruebas
	if req.Username != "admin" || req.Password != "admin123" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	user := models.User{
		Username: "admin",
		//Email:    "admin@library.com",
		Role: "admin",
	}

	token, err := auth.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// Me - Obtener información del usuario actual
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"role":    role,
	})
}
