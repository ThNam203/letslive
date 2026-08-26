package types

import "github.com/golang-jwt/jwt/v5"

type AdminClaims struct {
	AdminId string `json:"adminId"`
	jwt.RegisteredClaims
}
