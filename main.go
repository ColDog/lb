package main

import (
	"github.com/coldog/proxy/proxy"
)

func main() {
	config := map[string] interface{} {
		"key": "test",
		"ip_hash": false,
		"routes": []string{
			"/test",
			"/test/*",
		},
		"middleware": []string{},
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:8000", "health": "http://localhost:8000/check"},
			map[string] interface{} {"target": "http://localhost:8001", "health": "http://localhost:8001/check"},
			map[string] interface{} {"target": "http://localhost:8002", "health": "http://localhost:8002/check"},
			map[string] interface{} {"target": "http://localhost:8003", "health": "http://localhost:8003/check"},
		},
	}

	proxy := proxy.NewProxyServer(map[string] interface{} {
		"binds": "0.0.0.0",
		"port": 3000,
		"access_log": true,
	})

	proxy.Add(config)

	proxy.BuildRoutes()
	proxy.Start()
}