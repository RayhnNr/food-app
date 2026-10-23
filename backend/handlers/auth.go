package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"os"

	// "os/user"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100"`
	Username     string `json:"username" binding:"required,min=3,max=50"` // Tambahkan ini
	Email        string `json:"email" binding:"required,email,max=100"`
	Phone        string `json:"phone" binding:"required,min=10,max=20"`   // Tambahkan ini
	Password     string `json:"password" binding:"required,min=6,max=100"`
	CaptchaToken string `json:"captcha_token"`
}

// LoginRequest struktur untuk validasi input login
type LoginRequest struct {
	Identity     string `json:"identity" binding:"required"` 
	Password     string `json:"password" binding:"required"`
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
	if os.Getenv("APP_ENV") != "dev" {
		if !verifyCaptcha(req.CaptchaToken, c.ClientIP()) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"errors": map[string]string{"captcha": "Verifikasi captcha gagal"},
			})
			return
		}
	}
	

    identityClean := strings.ToLower(strings.TrimSpace(req.Identity))

    phoneSearch := req.Identity
    phoneSearch = strings.ReplaceAll(phoneSearch, " ", "")
    phoneSearch = strings.ReplaceAll(phoneSearch, "-", "")
    phoneSearch = strings.ReplaceAll(phoneSearch, "+", "")

    var user models.User
    err := database.DB.Where("email = ? OR username = ? OR phone = ?", identityClean, identityClean, phoneSearch).First(&user).Error
    
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "Authentication failed",
                "message": "Email, username, nomor telepon atau password salah",
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

	token, err := middlewares.GenerateToken(user.ID, user.Email, user.Name, user.Username, user.Role)
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
	if os.Getenv("APP_ENV") != "dev" {
		if !verifyCaptcha(req.CaptchaToken, c.ClientIP()) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"errors": map[string]string{"captcha": "Verifikasi captcha gagal"},
			})
			return
		}
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	
	phoneClean := req.Phone
	phoneClean = strings.ReplaceAll(phoneClean, " ", "")
	phoneClean = strings.ReplaceAll(phoneClean, "-", "")
	phoneClean = strings.ReplaceAll(phoneClean, "+", "")

	var existingUser models.User

	// 1. Cek duplikasi Username
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Username already registered",
			"message": "Username ini sudah terdaftar.",
		})
		return
	}

	// 2. Cek duplikasi Email
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Email already registered",
			"message": "Akun dengan email ini sudah terdaftar.",
		})
		return
	}

	// 3. Cek duplikasi Nomor Telepon
	if err := database.DB.Where("phone = ?", phoneClean).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Phone number already registered",
			"message": "Nomor telepon ini sudah terdaftar.",
		})
		return
	}

	user := models.User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Phone:    phoneClean,
		Password: req.Password,
		Role:     "user",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"message": "Gagal membuat akun user",
		})
		return // <- tambah
	}

	token, err := middlewares.GenerateToken(user.ID, user.Email, user.Name, user.Username, user.Role)
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
	userUsername, _ := c.Get("userUsername")
	userRole, _ := c.Get("userRole")  

	token, err := middlewares.GenerateToken(
		userID,
		userEmail.(string),
		userName.(string),
		userUsername.(string),
		userRole.(string),
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
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"username": user.Username,
			"email":    user.Email,
			"phone":    user.Phone,
			"role":     user.Role,
		},
	})
}

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "message": "API is running",
    })
}
