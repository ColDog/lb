package proxy

import (
	"net/http"
	"fmt"
)

type Context struct {
	finished	bool
	allowProxy	bool
	written 	bool
	Params 		map[string] string
	Writer 		http.ResponseWriter
	Handler 	Handler
	Req 		*http.Request
	Host		Host
}

func (ctx *Context) ClientIp() string {
	return ctx.Req.RemoteAddr
}

func (ctx *Context) Finish() {
	ctx.finished = true
	ctx.allowProxy = false
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

func (ctx *Context) WithStatus(status int) {
	result := fmt.Sprintf(`{"error": true, "code": %d, "message": "%s"}`, status, http.StatusText(status))
	ctx.Writer.WriteHeader(status)
	ctx.Write(result)
	ctx.AsJson()
	ctx.Finish()
}

func (ctx *Context) Write(body string) {
	ctx.written = true
	ctx.Writer.Write([]byte(body))
}

func (ctx *Context) SetHeader(key, value string) {
	ctx.Writer.Header().Add(key, value)
}

func (ctx *Context) AsJson() {
	ctx.SetHeader("Content-Type", "application/json")
}
