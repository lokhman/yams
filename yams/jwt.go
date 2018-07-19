package yams

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTClaims struct {
	jwt.StandardClaims
	Id int `json:"id"`

	Username string `json:"-"`
	Role     string `json:"-"`
}

func JWTSign(claims JWTClaims) string {
	claims.ExpiresAt = time.Now().Add(time.Hour).Unix()
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(SecretKey)
	if err != nil {
		panic(err)
	}
	return tokenString
}

func JWTParse(tokenString string, claims *JWTClaims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(`unexpected signing method "%v"`, token.Header["alg"])
		}
		return SecretKey, nil
	})
}
