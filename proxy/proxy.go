package proxy

import (
	"net/http"
	"fmt"
	"log"
	"github.com/coldog/proxy/tools"
	"github.com/coldog/proxy/router"
	"errors"
)

type BaseMiddleware func(ctx *Context) *Context

func NewProxyServer(config tools.Map) *ProxyServer {
	return &ProxyServer{
		binds: config.Str("binds"),
		port: config.Int32("port"),
		accessLog: config.Bool("access_log"),
		router: router.Router(),
		handlers: make(map[string] *Handler, 0),
		middleware: make(map[string] BaseMiddleware, 0),
	}
}

type ProxyServer struct {
	binds		string
	port		int32
	accessLog	bool
	handlers 	map[string] *Handler
	middleware	map[string] BaseMiddleware
	router 		*router.RouteTree
}

func (proxy ProxyServer) match(req *http.Request) (*Handler, map[string] string, bool) {
	key, params := proxy.router.Match(req.URL.Path)
	if handler, ok := proxy.handlers[key]; ok {
		return handler, params, true
	}

	return &Handler{}, params, false
}

func (proxy ProxyServer) handle(writer http.ResponseWriter, req *http.Request) {
	handler, params, ok := proxy.match(req)

	ctx := &Context{
		Writer: writer,
		Handler: *handler,
		Req: req,
		Params: params,
		finished: false,
		allowProxy: true,
	}

	var status string

	if ok {
		if host, ok := handler.Next(ctx); ok {
			ctx.Host = *host
			if handler.run(proxy, ctx) {
				status = "success"
				host.proxy.ServeHTTP(writer, req)
			} else {
				status = "halted"
			}
		} else {
			ctx.NoneAvailable()
			status = "no_hosts_available"
		}
	} else {
		ctx.NotFound()
		status = "no_handlers"
	}

	if proxy.accessLog {
		tools.Log("proxy.handled", map[string] interface{} {
			"key": "proxy.proxy",
			"status": status,
			"client_ip": ctx.ClientIp(),
			"handler_availability": handler.IsAvailable(),
			"handler_healthy_hosts": handler.GetAvailableHosts(),
			"handler_status": handler.Status(),
			"proxied_to": ctx.Host.target,
			"path": req.URL.Path,
			"server": fmt.Sprintf("%s:%d", proxy.binds, proxy.port),
		})
	}
}

func (proxy ProxyServer) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	proxy.handle(writer, req)
}

func (proxy ProxyServer) Start() {
	proxy.BuildRoutes()
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", proxy.binds, proxy.port), proxy))
}

func (proxy ProxyServer) Use(key string, f BaseMiddleware) {
	proxy.middleware[key] = f
}

func (proxy *ProxyServer) BuildRoutes() {
	router := router.Router()
	for _, handler := range proxy.handlers {
		for _, route := range handler.routes {
			router.Add(route, handler.key)
		}
	}

	proxy.router = router
}

func (proxy *ProxyServer) AddDefaultMiddleware() {
	proxy.Use("JwtAuth", JwtAuth)
	proxy.Use("JwtRedisAuth", JwtRedisAuth)
	proxy.Use("IpRateLimiter", IpRateLimiter)
}

func (proxy ProxyServer) Update(config tools.Map) error {
	key := config.Str("key")

	if key == "" {
		return errors.New("Must specify a key")
	}

	if _, ok := proxy.handlers[key]; !ok {
		proxy.handlers[key] = NewHandler(key)
	}

	proxy.handlers[key].Update(config)
	proxy.BuildRoutes()
	return nil
}

func (proxy *ProxyServer) UpdateAll(handlers tools.Map) error {
	if !handlers.Exists("handlers") {
		return errors.New("Must specify an array of handlers")
	}

	for _, config := range handlers.MapArray("handlers") {
		proxy.Update(config)
	}

	return nil
}

