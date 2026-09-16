package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sen1or/letslive/auth/domains"
)

func GenerateOTP() (string, error) {
	const otpLength = 6
	// exclusive
	maxOTPValue := big.NewInt(1000000)
	// random number from [0, maxOTPValue).
	n, err := rand.Int(rand.Reader, maxOTPValue)
	if err != nil {
		return "", domains.ErrFailedToCreateSignUpOTP
	}

	// Format the number as a 6-digit string, left-padding with zeros if needed.
	// For example:
	// n = 123    -> "000123"
	// n = 987654 -> "987654"
	// n = 0      -> "000000"
	otpFormat := fmt.Sprintf("%%0%dd", otpLength) // Creates the format string "%06d"
	otp := fmt.Sprintf(otpFormat, n)

	return otp, nil
}
