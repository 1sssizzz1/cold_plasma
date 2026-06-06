package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"
)

func NormalizeRussianPhone(phone string) (string, error) {
	var digits strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	v := digits.String()
	switch {
	case len(v) == 11 && v[0] == '8':
		v = "7" + v[1:]
	case len(v) == 11 && v[0] == '7':
	case len(v) == 10 && v[0] == '9':
		v = "7" + v
	default:
		return "", se(ErrValidation, "Укажите российский номер телефона")
	}
	if len(v) != 11 || v[0] != '7' || v[1] != '9' {
		return "", se(ErrValidation, "Укажите российский мобильный номер телефона")
	}
	return "+" + v, nil
}

func generatePhoneCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate phone code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func phoneCodeExpiresAt(ttlSeconds int) time.Time {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	return time.Now().Add(time.Duration(ttlSeconds) * time.Second)
}
