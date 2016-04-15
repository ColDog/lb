package main

import "github.com/coldog/proxy/proxy"

func main() {
	config := map[string] interface{} {
		"ip_hash": true,
		"path": "test",
		"middleware": []string{},
		"regex": "(.*)",
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:8000", "health": "http://localhost:8000/check"},
			map[string] interface{} {"target": "http://localhost:8001", "health": "http://localhost:8001/check"},
			map[string] interface{} {"target": "http://localhost:8002", "health": "http://localhost:8002/check"},
			map[string] interface{} {"target": "http://localhost:8003", "health": "http://localhost:8003/check"},
		},
	}

	proxy := proxy.NewProxyServer()

	proxy.Add(config)

	proxy.Start()
}