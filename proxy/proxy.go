package proxy

import (
	"net/http"
	"fmt"
	"log"
	"strings"
	"regexp"
)

type BaseMiddleware func(ctx *Context) *Context

type ProxyServer struct {
	binds		string
	workers		int32
	port		int32
	accessLog	bool
	handlers 	map[string] *Handler
	middleware	map[string] BaseMiddleware
}

func (proxy ProxyServer) match(req *http.Request) (*Handler, bool) {
	key := strings.Split(req.URL.Path, "/")[1]
	if handler, ok := proxy.handlers[key]; ok {
		return handler, true
	} else {
		for _, handler := range proxy.handlers {
			if handler.regex != "" {
				if match, _ := regexp.MatchString(handler.regex, "peach"); match {
					return handler, true
				}
			}
		}
	}

	return &Handler{}, false
}

func (proxy ProxyServer) handle(writer http.ResponseWriter, req *http.Request) {
	handler, ok := proxy.match(req)

	ctx := &Context{
		writer: writer,
		handler: *handler,
		req: req,
		finished: false,
		allowProxy: true,
	}

	if ok {
		if host, ok := handler.next(ctx); ok {
			ctx.host = host
			if proxy.run(handler, ctx) {
				if proxy.accessLog {
					log.Printf("[proxying]     PROXIED %s to %s", req.URL.Path, host.target)
				}
				host.proxy.ServeHTTP(writer, req)
			} else {
				if proxy.accessLog {
					log.Printf("[proxying]     HALTED  %s with %s", req.URL.Path, ctx.writer)
				}
			}
		} else {
			ctx.NoneAvailable()
			if proxy.accessLog {
				log.Printf("[proxying]     FAILED no host for %s handler: %s", req.URL.Path, ctx.handler.path)
			}
		}
	} else {
		ctx.NotFound()
		if proxy.accessLog {
			log.Printf("[proxying]     FAILED no handler for %s", req.URL.Path)
		}
	}
}

func (proxy ProxyServer) run(handler *Handler, ctx *Context) bool {
	for _, middleKey := range handler.middleware {
		if _, ok := proxy.middleware[middleKey]; ok {
			proxy.middleware[middleKey](ctx)
			if ctx.finished {
				break
			}
		}
	}

	return ctx.allowProxy
}

func (proxy ProxyServer) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	proxy.handle(writer, req)
}

func (proxy ProxyServer) Start() {
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", proxy.binds, proxy.port), proxy))
}

func (proxy ProxyServer) Use(key string, f BaseMiddleware) {
	proxy.middleware[key] = f
}

func (proxy ProxyServer) Add(config map[string] interface{}) error {
	key := config["path"].(string)

	if _, ok := proxy.handlers[key]; !ok {
		proxy.handlers[key] = &Handler{
			ip_hash: config["ip_hash"].(bool),
			middleware: config["middleware"].([]string),
			path: key,
			regex: config["regex"].(string),
			hosts: make([]*Host, 0),
			nextHost: 0,
			available: true,
		}

		proxy.handlers[key].StartHealthCheck()
	}

	proxy.handlers[key].update(config)
	log.Printf("Added Handler: %s", key)
	return nil
}

func (proxy *ProxyServer) Configure(handlers []map[string] interface{}) {
	for _, config := range handlers {
		proxy.Add(config)
	}
}

func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		binds: "0.0.0.0",
		port: 3000,
		workers: 10,
		accessLog: true,
		handlers: make(map[string] *Handler, 0),
		middleware: make(map[string] BaseMiddleware, 0),
	}
}
