package utils

import (
	"strings"
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	otp, err := GenerateOTP()
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}
	if len(otp) != OTPLength {
		t.Fatalf("GenerateOTP() length = %d, want %d", len(otp), OTPLength)
	}
	for _, char := range otp {
		if !strings.ContainsRune(otpCharSet, char) {
			t.Fatalf("GenerateOTP() returned unsupported character %q", char)
		}
	}
}

func TestHashAndVerifyOTP(t *testing.T) {
	const otp = "ABCDEFGH23"
	hash, err := HashOTP(otp)
	if err != nil {
		t.Fatalf("HashOTP() error = %v", err)
	}
	if !VerifyOTP(hash, otp) {
		t.Fatal("VerifyOTP() should accept the matching OTP")
	}
	if VerifyOTP(hash, "ABCDEFGH24") {
		t.Fatal("VerifyOTP() accepted an incorrect OTP")
	}
}
