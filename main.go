package main

import (
	"net/http"
	"github.com/coldog/proxy/proxy"
	"github.com/coldog/proxy/tools"
	"flag"
	"fmt"
)

type App struct {
	proxy 	proxy.ProxyServer
}

func main() {
	host := flag.String("server-host", "0.0.0.0", "Host to bind to")
	port := flag.Int("server-port", 8080, "Port to bind to")

	proxyHost := flag.String("proxy-host", "0.0.0.0", "Host to bind to")
	proxyPort := flag.Int("proxy-port", 3000, "Port to bind to")
	accessLog := flag.Bool("access_log", true, "Access log enabled")

	app := &App{proxy.NewProxyServer(tools.NewMap(map[string] interface{} {
		"binds": proxyHost,
		"port": proxyPort,
		"access_log": accessLog,
	}))}

	go func() {
		app.proxy.BuildRoutes()
		app.proxy.Start()
	}()

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), app)
}
