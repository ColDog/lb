package ctx

import (
	"fmt"
	"net/http"
)

func NewCtx(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
	}
}

type Context struct {
	Finished bool
	Written  bool
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
	ctx.Write(result)
	ctx.AsJson()
	ctx.Finish()
}

func (ctx *Context) Write(body string) {
	ctx.Written = true
	ctx.Writer.Write([]byte(body))
}

func (ctx *Context) SetHeader(key, value string) {
	ctx.Writer.Header().Add(key, value)
}

func (ctx *Context) AsJson() {
	ctx.SetHeader("Content-Type", "application/json")
}
