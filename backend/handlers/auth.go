package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	// "os/user"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	CaptchaToken string `json:"captcha_token"`
}

// LoginRequest struktur untuk validasi input login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	CaptchaToken string `json:"captcha_token"`
}

type AuthResponse struct {
	Message string `json:"message"`
	User models.PublicUser `json:"user"`
	Token string `json:"token"`
}


func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return // <- tambah
	}
	if !verifyCaptcha(req.CaptchaToken, c.ClientIP()) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{"captcha": "Verifikasi captcha gagal"},
		})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Email atau password salah",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"message": "Gagal ambil data user",
			})
		}
		return
	}

	if err := user.CheckPassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": "Email atau password salah", // <- disamakan
		})
		return
	}

	token, err := middlewares.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Login berhasil tapi gagal generate auth token",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{ // <- 200, bukan 201
		Message: "Login berhasil",
		User:    user.ToPublicUser(),
		Token:   token,
	})
}

func Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return // <- tambah
	}

	if !verifyCaptcha(req.CaptchaToken, c.ClientIP()) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{"captcha": "Verifikasi captcha gagal"},
		})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	var existingUser models.User
	result := database.DB.Where("email = ?", req.Email).First(&existingUser)

	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Email already registered",
			"message": "Akun dengan email ini sudah terdaftar.",
		})
		return
	} else if result.Error != gorm.ErrRecordNotFound { // <- tambah
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"message": "Gagal memeriksa email",
		})
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // di-hash oleh hook BeforeCreate di model
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"message": "Gagal membuat akun user",
		})
		return // <- tambah
	}

	token, err := middlewares.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Akun sudah dibuat tapi gagal generate auth token",
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "Registrasi berhasil",
		User:    user.ToPublicUser(),
		Token:   token,
	})
}

func RefreshTOken(c *gin.Context) {
	// ambil user info dari context
	userID, _ := middlewares.GetUserIDFromCOntext(c)
	userEmail, _ := c.Get("userEmail")
	userName, _ := c.Get("userName")

	token, err := middlewares.GenerateToken(
		userID,
		userEmail.(string),
		userName.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Token generation failed",
			"message" : "Gagal dalam refresh token auth",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message" : "Token berhasil di refresh",
		"token" : token,
	})
}

func GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{"id": user.ID, "name": user.Name, "email": user.Email},
	})
}

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "message": "API is running",
    })
}
