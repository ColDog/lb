package proxy

import (
	"testing"
	"net/http"
)

func TestHandler(t *testing.T) {
	config := map[string] interface{} {
		"ip_hash": true,
		"path": "test",
		"middleware": []string{"json"},
		"regex": "(.*)",
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:3000", "health": "http://localhost:3001"},
		},
	}

	proxy := ProxyServer{handlers: make(map[string] *Handler, 0)}

	proxy.Add(config)

	ctx := &Context{
		writer: MockWriter{},
		req: &http.Request{},
		finished: false,
		allowProxy: true,
	}

	host, ok := proxy.handlers["test"].next(ctx)
	if !ok {
		t.Fail()
	}

	if host.target != "http://localhost:3000" {
		t.Fail()
	}
}
