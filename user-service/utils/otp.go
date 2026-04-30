package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

// GenerateOTP generates a 6-digit OTP code
func GenerateOTP() (string, error) {
	const digits = "0123456789"
	var otp string

	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			log.Printf("Error generating OTP: %v", err)
			return "", err
		}
		otp += string(digits[num.Int64()])
	}

	return otp, nil
}

// FormatOTPMessage creates a nice email message for OTP
func FormatOTPMessage(name, otp string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 20px; border-radius: 5px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; border-radius: 5px; margin: 20px 0; }
        .otp-box { background: white; border: 2px solid #667eea; padding: 15px; border-radius: 5px; text-align: center; margin: 20px 0; }
        .otp-code { font-size: 32px; font-weight: bold; color: #667eea; letter-spacing: 5px; }
        .footer { font-size: 12px; color: #666; text-align: center; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to PaymentApp! 💳</h1>
        </div>
        <div class="content">
            <p>Hi <strong>%s</strong>,</p>
            <p>Thank you for signing up! Please use the verification code below to complete your account creation:</p>
            <div class="otp-box">
                <p>Your Verification Code:</p>
                <div class="otp-code">%s</div>
                <p style="color: #999; font-size: 12px;">This code expires in 10 minutes</p>
            </div>
            <p><strong>Important:</strong> Never share this code with anyone. Our team will never ask for your OTP.</p>
            <p>If you didn't create this account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 PaymentApp. All rights reserved.</p>
            <p>Secure • Fast • Reliable</p>
        </div>
    </div>
</body>
</html>`,
		name, otp,
	)
}
