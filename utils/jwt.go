package utils

import (
	"time"

	"aide-devoir-forum/models"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWTToken génère un token JWT pour un utilisateur
func GenerateJWTToken(userID int, username string, roleID int, secretKey []byte, expirationTime time.Duration) (string, error) {
	claims := &models.Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseJWTToken parse et valide un token JWT
func ParseJWTToken(tokenString string, secretKey []byte) (*models.Claims, error) {
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// ValidateJWTToken vérifie si un token JWT est valide
func ValidateJWTToken(tokenString string, secretKey []byte) bool {
	_, err := ParseJWTToken(tokenString, secretKey)
	return err == nil
}
