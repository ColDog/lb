package proxy

import (
	"testing"
	"net/http"
	"github.com/coldog/proxy/tools"
)

func TestHandler(t *testing.T) {
	config := tools.NewMap(map[string] interface{} {
		"ip_hash": true,
		"key": "test",
		"routes": []string{},
		"middleware": []string{"json"},
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:3000", "health": "http://localhost:3001"},
		},
	})

	proxy := ProxyServer{handlers: make(map[string] *Handler, 0)}

	proxy.Update(config)

	ctx := &Context{
		Writer: MockWriter{},
		Req: &http.Request{},
		finished: false,
		allowProxy: true,
	}

	host, ok := proxy.handlers["test"].Next(ctx)
	if !ok {
		t.Fail()
	}

	if host.target != "http://localhost:3000" {
		t.Fail()
	}
}
