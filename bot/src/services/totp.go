package services

import (
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateOTP(user string) (string, error) {
	totp, err := totp.GenerateCodeCustom(os.Getenv("OTP_SECRET"), time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA512,
	})

	if err != nil {
		return "", err
	}

	KV.Set(user+":totp", totp, time.Minute*5)

	return totp, nil
}

func ValidateOTP(user string, otp string) bool {
	sk := KV.Get(user + ":totp").Value.(string)

	return totp.Validate(otp, sk)
}
