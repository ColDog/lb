package proxy

import "testing"

func TestSimple(t *testing.T) {
	config := map[string] interface{} {
		"ip_hash": true,
		"routes": []string{"test"},
		"key": "test",
		"middleware": []string{"json"},
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:3000", "health": "http://localhost:3001"},
		},
	}

	proxy := ProxyServer{
		handlers: make(map[string] *Handler, 0),
		middleware: make(map[string] BaseMiddleware, 0),
	}

	proxy.Add(config)

	proxy.Use("test", func(ctx *Context) *Context {
		return ctx
	})
}
