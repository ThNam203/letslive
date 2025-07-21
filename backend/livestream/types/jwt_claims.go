package types

import "github.com/golang-jwt/jwt/v5"

type MyClaims struct {
	UserId   string `json:"userId"`
	Consumer string `json:"consumer"`
	jwt.RegisteredClaims
}
