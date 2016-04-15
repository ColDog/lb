package proxy

import (
	"log"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"os"
)

func Logger(ctx *Context) (*Context) {
	ctx.writer.Header().Add("X-Test", "hello there")
	log.Printf("I am some logging middleware! request %s\n", ctx.handler.path)
	return ctx
}

func JwtAuth(ctx *Context) (*Context) {
	authHeader := strings.Split(ctx.req.Header["Authorization"][0], " ")[1]
	token, err := jwt.Parse(authHeader, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_SECRET"), nil
	})
	if err != nil || !token.Valid {
		ctx.Unauthorized()
	}
	return ctx
}

func JwtRedisAuth(ctx *Context) (*Context) {
	authHeader := strings.Split(ctx.req.Header["Authorization"][0], " ")[1]
	token, err := jwt.Parse(authHeader, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_SECRET"), nil
	})
	if err != nil || !token.Valid {
		ctx.Unauthorized()
	}
	return ctx
}
