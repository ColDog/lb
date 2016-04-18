package proxy

import (
	"net/http"
	"time"
	"net/url"
	"net/http/httputil"
	"hash/fnv"
	"github.com/coldog/proxy/tools"
)

// The data structure and methods for handlers and hosts.
// handlers have a list of hosts that can respond to requests, as well as a list of middleware that the handler should
// step through for each request. Handlers are mutable, hosts are immutable. Configuration is best handled by increasing
// versions of a json file, where a handler is defined solely by a single json file. This can be mapped to the 'key' in
// the handler struct.
//
// Example configuration:
// {
//	"key": "test",
//	"ip_hash": false,
//	"routes": [
// 		"/test",
// 		"/test/*"
// 	],
//	"middleware": [
// 		"JwtAuth"
// 	],
//	"hosts": [
//		{
// 			"target": "http://localhost:8000",
// 			"health": "http://localhost:8000/health",
//			"timeout": 10,
//			"down": false
// 		}
// 	]
// }
//
// The handler Update method, would handle this json structure to update the handler accordingly. It would create a new
// set of hosts and add them to the list of hosts and update the other configuration. The handler struct is threadsafe which
// allows concurrent updates.

type Host struct {
	target		string
	health		string
	timeout 	int
	down 		bool
	proxy 		http.Handler
	checked		int64
	healthy		bool
}

// Checks whether this host is down and marked as healthy by the healthchecker.
func (host *Host) available() bool {
	return host.healthy && !host.down
}

// Pings to mark the host as healthy or not.
func (host *Host) ping() {
	if host.timeout < 5 {
		host.timeout = 5
	}
	timeout := time.Duration(time.Duration(host.timeout) * time.Second)
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(host.health)
	host.healthy = err == nil && resp.StatusCode == 200
	host.checked = time.Now().Unix()
}

// Create a new default handler. Only takes key as an argument. Use the update method on this new handler to configure.
func NewHandler(key string) *Handler {
	handler := &Handler{
		key: key,
		middleware: make([]string, 0),
		hosts: make([]*Host, 0),
		nextHost: 0,
		available: true,
	}
	handler.StartHealthCheck()
	return handler
}

type Handler struct {
	ip_hash 	bool
 	routes 		[]string
	key 		string
	hosts		[]*Host
	middleware	[]string
	available	bool
	down 		bool
	nextHost	int
	healthyHosts    []int
}

// Returns the percentage of available hosts for a given handler.
func (handler *Handler) Status() float64 {
	return  float64(len(handler.GetAvailableHosts())) / float64(len(handler.hosts))
}

// Get a list of the available hosts. Returns the indices.
func (handler *Handler) GetAvailableHosts() []int {
	if handler.healthyHosts == nil {
		handler.healthyHosts = make([]int, 0)
		for idx, host := range handler.hosts {
			if host.available() {
				handler.healthyHosts = append(handler.healthyHosts, idx)
			}
		}
	}

	return handler.healthyHosts
}

// This is the load balancing method. It returns the host to be used for a given request as well as an 'ok' flag to let
// you know if it could find any available hosts. The reverse proxy should return 503 if false is retuned as the second
// argument.
func (handler *Handler) Next(ctx *Context) (*Host, bool) {
	idx := handler.nextHost
	if idx >= len(handler.hosts) {
		idx = 0
	}

	if !handler.IsAvailable() {
		return &Host{}, false
	}

	if len(handler.hosts) == 1 {
		handler.nextHost = 0
		return handler.hosts[0], true
	}

	healthy := handler.GetAvailableHosts()
	if len(healthy) == 1 {
		handler.nextHost = healthy[0]
		return handler.hosts[handler.nextHost], true
	}

	if handler.ip_hash {
		ip := ctx.ClientIp()
		if ip != "" && ip != "unknown" {
			// build a map of the result to the healthy hosts
			h := fnv.New64a()
			h.Write([]byte(ip))
			key := h.Sum64()

			var b, j int64
			for j < int64(len(handler.hosts) -1) {
				b = j
				key = key*2862933555777941757 + 1
				j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
			}

			handler.nextHost = int(b)
			if handler.hosts[int(b)].available() {
				return handler.hosts[int(b)], true
			}
		}
	}

	if len(handler.hosts) > 1 {

		for ; idx < len(handler.hosts); {
			if !handler.hosts[idx].available() {
				idx++
			}

			if handler.hosts[idx].available() {
				break
			}

			idx++
		}

		handler.nextHost = idx + 1
		return handler.hosts[idx], true
	}

	tools.Log("handler", map[string] interface{} {
		"event": "proxy.next_host",
		"status": "no_hosts_available",
		"handler": handler.key,
	})
	return &Host{}, false
}

// steps through the middleware for a given proxy server for this handler.
func (handler *Handler) run(proxy ProxyServer, ctx *Context) bool {
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

// starts the health check loop.
func (handler *Handler) StartHealthCheck() {
	go func() {
		for  {

			available := false
			healthy := make([]int, 0)
			for idx, host := range handler.hosts {
				host.ping()
				if host.available() {
					healthy = append(healthy, idx)
					available = true
				}
			}

			handler.healthyHosts = healthy
			handler.available = available

			tools.Log("healthcheck", map[string] interface{} {
				"key": "proxy.healthcheck",
				"available": handler.IsAvailable(),
				"healthy_hosts": handler.GetAvailableHosts(),
				"status": handler.Status(),
			})
			time.Sleep(15 * time.Second)
		}
	}()
}

// Updates the handler with the latest configuration. Uses a wrapped Map from the tools package. This just handles some
// of the more tedious type checking. With a json structure, it can be instantiated with:
// 	handler.Update(tools.Map{json: []bytes(`{"x": "y"}`)
// Concurrency is handled by readers making sure that they
func (handler *Handler) Update(config tools.Map) {
	handler.middleware = config.StrArray("middleware")
	handler.ip_hash = config.Bool("ip_hash")
	handler.routes = config.StrArray("routes")

	newHosts := make([]*Host, 0)
	for _, host := range config.MapArray("hosts") {

		dest, _ := url.Parse(host.Str("target"))
		proxy := httputil.NewSingleHostReverseProxy(dest)

		hostStruct := &Host{
			health: host.Str("health"),
			target: host.Str("target"),
			timeout: host.Int("timeout"),
			down: host.Bool("down"),
			checked: 0,
			healthy: true,
			proxy: proxy,
		}

		newHosts = append(newHosts, hostStruct)
	}

	handler.hosts = newHosts
}

// Marks the current handler as down
func (handler *Handler) MarkDown() {
	handler.down = true
}

// Marks the current handler as up
func (handler *Handler) MarkUp() {
	handler.down = false
}

// Checks if not down and is marked available by the health check loop.
func (handler *Handler) IsAvailable() bool {
	return handler.available && !handler.down
}
