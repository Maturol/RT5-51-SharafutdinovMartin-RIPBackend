package ds

import "github.com/golang-jwt/jwt"

type JWTClaims struct {
	jwt.StandardClaims
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	IsModerator bool   `json:"is_moderator"`
}
