package utils

import "time"

const (
	authTokenExp              = time.Minute * 10
	refreshTokenExp           = time.Hour * 24 * 30
	blacklistTokenExp         = "blacklisted:"
	OTPExpiration             = time.Minute * 10
	otpCharSet                = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	passwordResetEmailSubject = "Password reset code"
	emailTemplate             = "Your password reset code is %s. It expires in %.0f minutes.\r\n" +
		"If you did not request this, you can ignore this email.\r\n"

	OTPLength = 10
)
