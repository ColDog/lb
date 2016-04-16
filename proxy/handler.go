package proxy

import (
	"net/http"
	"time"
	"log"
	"net/url"
	"net/http/httputil"
	"hash/fnv"
	"errors"
)

type Host struct {
	target		string
	health		string
	proxy 		http.Handler
	checked		int64
	healthy		bool
}

func (host *Host) ping() {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(host.health)
	host.healthy = err == nil && resp.StatusCode == 200
	host.checked = time.Now().Unix()
}

type Handler struct {
	ip_hash 	bool
	regex  		string
 	path 		string
	hosts		[]*Host
	middleware	[]string
	available	bool
	nextHost	int
	healthyHosts    []int
}


func (handler *Handler) update(config map[string] interface{}) {
	handler.middleware = config["middleware"].([]string)
	handler.ip_hash = config["ip_hash"].(bool)
	handler.middleware = config["middleware"].([]string)
	handler.regex = config["regex"].(string)

	for _, host := range config["hosts"].([]map[string] interface{}) {
		if !handler.hasHost(host["target"].(string)) {
			handler.add(host)
		}
	}
}

func (handler *Handler) hasHost(url string) bool {
	for _, host := range handler.hosts {
		if host.target == url {
			return true
		}
	}

	return false
}

func (handler *Handler) Status() float64 {
	return  float64(len(handler.getHealthyHosts())) / float64(len(handler.hosts))
}

func (handler *Handler) getHealthyHosts() []int {
	if handler.healthyHosts == nil {
		handler.healthyHosts = make([]int, 0)
		for idx, host := range handler.hosts {
			if host.healthy {
				handler.healthyHosts = append(handler.healthyHosts, idx)
			}
		}
	}

	return handler.healthyHosts
}

func (handler *Handler) next(ctx *Context) (*Host, bool) {
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

	healthy := handler.getHealthyHosts()
	if len(healthy) == 1 {
		handler.nextHost = healthy[0]
		return handler.hosts[handler.nextHost], true
	}

	if handler.ip_hash {
		ip, ok := ctx.clientIp()
		if ok {
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
			if handler.hosts[int(b)].healthy {
				return handler.hosts[int(b)], true
			}
		}
	}

	if len(handler.hosts) > 1 {

		for ; idx < len(handler.hosts); {
			if !handler.hosts[idx].healthy {
				idx++
			}

			if handler.hosts[idx].healthy {
				break
			}

			idx++
		}

		handler.nextHost = idx + 1
		return handler.hosts[idx], true
	}

	log.Printf("[handler]      %s failed to find a suitable host", handler.path)
	return &Host{}, false
}

func (handler *Handler) add(config map[string] interface{}) error {
	if handler.hasHost(config["target"].(string)) {
		log.Println("already has host")
		return errors.New("Already has that host")
	}

	dest, _ := url.Parse(config["target"].(string))
	proxy := httputil.NewSingleHostReverseProxy(dest)

	hostStruct := &Host{
		health: config["health"].(string),
		target: config["target"].(string),
		checked: 0,
		healthy: true,
		proxy: proxy,
	}

	handler.hosts = append(handler.hosts, hostStruct)
	return nil
}

func (handler *Handler) StartHealthCheck() {
	go func() {
		for  {

			available := false
			healthy := make([]int, 0)
			for idx, host := range handler.hosts {
				host.ping()
				if host.healthy {
					healthy = append(healthy, idx)
					available = true
				}
			}

			handler.healthyHosts = healthy
			handler.available = available

			log.Printf("[healthcheck]  %s available: %t, healthy list: %d, up percent: %f", handler.path, handler.IsAvailable(), handler.getHealthyHosts(), handler.Status())

			time.Sleep(15 * time.Second)
		}
	}()
}

func (handler *Handler) MarkUnvailable() {
	handler.available = false
}

func (handler *Handler) MarkAvailable() {
	handler.available = true
}

func (handler *Handler) IsAvailable() bool {
	return handler.available
}
