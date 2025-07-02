package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

type JwtCustomClaims struct {
	PatientID uint   `json:"patient_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(patientID uint, email string) (string, error) {
	claims := &JwtCustomClaims{
		PatientID: patientID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return t, nil
}
