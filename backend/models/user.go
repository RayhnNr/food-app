package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"size:100;not null"`
	Email    string `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"` // json:"-" = tidak pernah dikirim ke klien
	Role     string `json:"role" gorm:"size:20;default:user"`
	CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// dipanggil sebelum create
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if len(u.Password) < 6 {
		return errors.New("Password minimal 6 Karakter")
	}

	hashedPassword, err := u.HashPassword(u.Password)

	if err != nil {
		 return err
	}

	u.Password = hashedPassword

	return nil


}

// enkripsi password menggunakan bcrypt
func (u *User) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Cek password dengan hash yang tersimpan
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// Buat response API tanpa Password
type PublicUser struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
	Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToPublicUser() PublicUser {
	return PublicUser{
		ID: u.ID,
		Name: u.Name,
		Email: u.Email,
		Role: u.Role,
		CreatedAt: u.CreatedAt,
	}
}

