package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	// "go/token"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"` // Tambahkan ini
	Role     string `json:"role"`     // Tambahkan ini
	jwt.RegisteredClaims
}


// Generate JWT Token baru untuk user
func GenerateToken(userID uint, email string, name string, username string, role string) (string, error) {
	// Waktu Expired token
	expirationTime := time.Now().Add(24 * time.Hour)

	// membuat claim (payload) untuk token
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Name:     name,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-gin-auth-api",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	// Membua token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET belum di set")
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Authmiddleware -> verifikasi JWT token dari request header
func AuthMiddleware() gin.HandlerFunc {
	return  func(c *gin.Context) {
		// Ambil token dari header auth
		authHeader := c.GetHeader("Authorization")

		// validasi format
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error" : "Authorization header is required",
				"message" : "Harap beirkan token JWT valid pada header",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error" : "Invalid authorization format",
				"message" : "Header Authorization harus menggunakan format: Bearer <token>",
			})
		}

		tokenString := tokenParts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Signing method tidak sesuai dengan yang diharapkan: %v", token.Header["alg"])
			}

			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				return nil, fmt.Errorf("JWT_SECRET belum dikonfigurasikan")
			}

			return []byte(jwtSecret), nil
		})

		// Handle berbagai error case
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid Token Signature",
					"message": "Tanda tangan token tidak valid atau telah dimodifikasi.",
				})
				return
			} else if err == jwt.ErrTokenExpired {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error" : "Token expired",
					"message" : "Session Expired, tolong Login kembali",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error" : "Invalid Token",
					"Message" : "Token yang diberikan tidak valid",
				})
			}
			c.Abort()
			return 
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error" : "Invalid token",
				"message" : "Token tidak valid",
			})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userName", claims.Name)
		c.Set("userUsername", claims.Username) // Tambahkan ini
		c.Set("userRole", claims.Role)  
		c.Set("claims", claims)

		// Lanjur ke hendler berikutnya
		c.Next()
	}
}

func GetUserIDFromCOntext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	if id, ok := userID.(uint); ok {
		return  id, true
	}

	return 0, false
}