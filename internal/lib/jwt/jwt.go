package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kurochkinivan/auth/internal/entity"
)

// TODO: покрыть тестами
func NewToken(user *entity.User, secret string, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(ttl).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
