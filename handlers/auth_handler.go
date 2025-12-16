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

// Register - Registrar nuevo usuario (VERSIÓN CORREGIDA Y UNIFICADA)
// Register - Registrar nuevo usuario
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest // ← Ahora solo tiene Username y Password

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar si usuario ya existe
	existingUser, err := h.store.GetUserByUsername(req.Username)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing password"})
		return
	}

	// Crear usuario (sin Email ni FullName)
	user := models.User{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guardar usuario en el store
	createdUser, err := h.store.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user: " + err.Error()})
		return
	}

	// Generar token
	token, err := auth.GenerateToken(createdUser.Username, createdUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// No devolver la contraseña
	createdUser.Password = ""

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  *createdUser,
	})
}

// Login - Iniciar sesión
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buscar usuario
	user, err := h.store.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// TEMPORAL: Comparar contraseñas SIN bcrypt
	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generar token
	token, err := auth.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// No devolver password
	user.Password = ""

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  *user,
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
