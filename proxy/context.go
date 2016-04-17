package proxy

import (
	"net/http"
	"fmt"
	"strings"
)

var IpHeaders []string = []string {
	"X-Forwarded-For",
	"Proxy-Client-IP",
	"WL-Proxy-Client-IP",
	"HTTP_X_FORWARDED_FOR",
	"HTTP_X_FORWARDED",
	"HTTP_X_CLUSTER_CLIENT_IP",
	"HTTP_CLIENT_IP",
	"HTTP_FORWARDED_FOR",
	"HTTP_FORWARDED",
	"HTTP_VIA",
	"REMOTE_ADDR",
};

type Context struct {
	finished	bool
	allowProxy	bool
	written 	bool
	Params 		map[string] string
	Writer 		http.ResponseWriter
	Handler 	Handler
	Req 		*http.Request
	Host		*Host
}

func (ctx *Context) ClientIp() (string, bool) {
	for _, header := range IpHeaders {
		if ip, ok := ctx.Req.Header[header]; ok && len(ip) >= 1 && strings.ToLower(ip[0]) != "unknown" && ip[0] != "" {
			return ip[0], true
		}
	}

	return "", false
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
