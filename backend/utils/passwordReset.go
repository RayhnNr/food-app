package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"food-api/config"
	"golang.org/x/crypto/bcrypt"
)

func GenerateOTP() (string, error) {
	result := make([]byte, OTPLength)
	charsetLength := big.NewInt(int64(len(otpCharSet)))

	for i := range result {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("generate reset OTP: %w", err)
		}
		result[i] = otpCharSet[num.Int64()]
	}

	return string(result), nil
}

func HashOTP(otp string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash reset OTP: %w", err)
	}
	return string(hash), nil
}

func VerifyOTP(hash, otp string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(otp)) == nil
}

func SendOTP(otp string, recipient string) error {
	body := fmt.Sprintf(emailTemplate, otp, OTPExpiration.Minutes())
	return config.SendEmail(recipient, passwordResetEmailSubject, body)
}
