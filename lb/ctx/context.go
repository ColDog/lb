package ctx

import (
	"fmt"
	"net/http"
)

func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Quit:   make(chan struct{}),
	}
}

type Context struct {
	Finished bool
	Writer   http.ResponseWriter
	Req      *http.Request
	Quit     chan struct{}
}

func (ctx *Context) ClientIp() string {
	return ctx.Req.RemoteAddr
}

func (ctx *Context) Unauthorized() {
	ctx.WithStatus(401)
}

func (ctx *Context) Forbidden() {
	ctx.WithStatus(403)
}

func (ctx *Context) NotFound() {
	ctx.WithStatus(404)
}

func (ctx *Context) NoneAvailable() {
	ctx.WithStatus(503)
}

func (ctx *Context) Finish() {
	ctx.Finished = true
	close(ctx.Quit)
}

func (ctx *Context) WithStatus(status int) {
	result := fmt.Sprintf(`{"error": true, "code": %d, "message": "%s"}`, status, http.StatusText(status))
	ctx.Writer.WriteHeader(status)
	ctx.AsJson()
	ctx.Write(result)
}

func (ctx *Context) Write(body string) {
	ctx.Writer.Write([]byte(body))
	ctx.Finish()
}

func (ctx *Context) SetHeader(key, value string) {
	ctx.Writer.Header().Add(key, value)
}

func (ctx *Context) AsJson() {
	ctx.SetHeader("Content-Type", "application/json")
}
