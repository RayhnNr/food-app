package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"food-api/utils"
	"log"
	"net/http"
	"os"
	"time"

	// "os/user"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RegisterRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100"`
	Username     string `json:"username" binding:"required,min=3,max=50"` // Tambahkan ini
	Email        string `json:"email" binding:"required,email,max=100"`
	Phone        string `json:"phone" binding:"required,min=10,max=20"` // Tambahkan ini
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
	Message string            `json:"message"`
	User    models.PublicUser `json:"user"`
	Token   string            `json:"token"`
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
			"error":   "Token generation failed",
			"message": "Gagal dalam refresh token auth",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token berhasil di refresh",
		"token":   token,
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
		"status":  "healthy",
		"message": "API is running",
	})
}

// Lupa Password

func SendPasswordResetEmail(c *gin.Context) {
	type Input struct {
		Email string `json:"email" binding:"required,email,max=100"`
	}

	var input Input

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": "Masukkan alamat email yang valid",
		})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	response := gin.H{
		"message": "Jika email terdaftar, kode reset akan dikirim.",
	}

	var user models.User
	err := database.DB.Where("email = ?", input.Email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusAccepted, response)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses permintaan reset password"})
		return
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kode reset password"})
		return
	}
	otpHash, err := utils.HashOTP(otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kode reset password"})
		return
	}

	now := time.Now().UTC()
	reset := models.PasswordResetOTP{
		UserID:    user.ID,
		OTPHash:   otpHash,
		ExpiresAt: now.Add(utils.OTPExpiration),
		CreatedAt: now,
	}
	cooldown := false
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var lockedUser models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedUser, user.ID).Error; err != nil {
			return err
		}

		var existing models.PasswordResetOTP
		err := tx.Where("user_id = ?", lockedUser.ID).First(&existing).Error
		if err == nil && now.Sub(existing.CreatedAt) < time.Minute {
			cooldown = true
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		reset.UserID = lockedUser.ID
		if err := tx.Where("user_id = ?", lockedUser.ID).Delete(&models.PasswordResetOTP{}).Error; err != nil {
			return err
		}
		return tx.Create(&reset).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kode reset password"})
		return
	}
	if cooldown {
		c.JSON(http.StatusAccepted, response)
		return
	}

	if err := utils.SendOTP(otp, user.Email); err != nil {
		log.Printf("password reset email delivery failed: %v", err)
		if deleteErr := database.DB.Delete(&reset).Error; deleteErr != nil {
			log.Printf("failed to remove undelivered password reset OTP: %v", deleteErr)
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email reset password tidak dapat dikirim saat ini"})
		return
	}

	c.JSON(http.StatusAccepted, response)
}

func ResetPassword(c *gin.Context) {
	type Input struct {
		Email       string `json:"email" binding:"required,email,max=100"`
		OTP         string `json:"otp" binding:"required,len=10"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": "Email, kode reset, dan password baru tidak valid",
		})
		return
	}
	if len([]byte(input.NewPassword)) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": "Password baru maksimal 72 byte",
		})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.OTP = strings.ToUpper(strings.TrimSpace(input.OTP))
	now := time.Now().UTC()
	invalidCode := false
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("email = ?", input.Email).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				invalidCode = true
				return nil
			}
			return err
		}

		var reset models.PasswordResetOTP
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND expires_at > ? AND used_at IS NULL AND attempts < ?", user.ID, now, 5).
			First(&reset).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				invalidCode = true
				return nil
			}
			return err
		}

		if !utils.VerifyOTP(reset.OTPHash, input.OTP) {
			invalidCode = true
			return tx.Model(&reset).UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
		}

		hashedPassword, err := user.HashPassword(input.NewPassword)
		if err != nil {
			return err
		}
		if err := tx.Model(&user).Update("password", hashedPassword).Error; err != nil {
			return err
		}
		return tx.Model(&reset).Update("used_at", now).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mereset password"})
		return
	}
	if invalidCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode reset tidak valid atau sudah kedaluwarsa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diperbarui"})
}
