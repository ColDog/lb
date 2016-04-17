package proxy

import (
	"strings"
	"github.com/dgrijalva/jwt-go"
	"os"
)

func JwtAuth(ctx *Context) *Context {
	authHeader := strings.Split(ctx.Req.Header["Authorization"][0], " ")[1]
	token, err := jwt.Parse(authHeader, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_SECRET"), nil
	})
	if err != nil || !token.Valid {
		ctx.Unauthorized()
	}
	return ctx
}

func JwtRedisAuth(ctx *Context) *Context {
	authHeader := strings.Split(ctx.Req.Header["Authorization"][0], " ")[1]
	token, err := jwt.Parse(authHeader, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_SECRET"), nil
	})
	if err != nil || !token.Valid {
		ctx.Unauthorized()
	}
	return ctx
}

func IpRateLimiter(ctx *Context) *Context  {
	return ctx
}
