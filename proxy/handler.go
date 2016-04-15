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
	lastHost	int
	healthyHosts    []int
	healthyCount 	int
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
		handler.healthyCount = -1 // to invalidate cache
	}

	if handler.healthyCount == len(handler.healthyHosts) {
		return handler.healthyHosts
	} else {
		handler.healthyHosts = make([]int, 0)
		for idx, host := range handler.hosts {
			if host.healthy {
				handler.healthyHosts = append(handler.healthyHosts, idx)
			}
		}
		handler.healthyCount = len(handler.healthyHosts)
	}

	return handler.healthyHosts
}

func (handler *Handler) next(ctx *Context) (*Host, bool) {
	if !handler.IsAvailable() {
		return &Host{}, false
	}

	if len(handler.hosts) == 1 {
		handler.lastHost = 0
		return handler.hosts[handler.lastHost], true
	}

	healthy := handler.getHealthyHosts()
	if len(healthy) == 1 {
		handler.lastHost = healthy[0]
		return handler.hosts[handler.lastHost], true
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

			handler.lastHost = int(b)
			if handler.hosts[handler.lastHost].healthy {
				return handler.hosts[handler.lastHost], true
			}
		}
	}

	if len(handler.hosts) > 1 {
		handler.lastHost++

		if handler.lastHost >= len(handler.hosts) {
			handler.lastHost = 0
		}

		for ; handler.lastHost < len(handler.hosts); handler.lastHost++ {
			if handler.hosts[handler.lastHost].healthy {
				break
			}
		}

		return handler.hosts[handler.lastHost], true
	}

	log.Printf("[%s] failed to find a suitable host", handler.path)
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
			count := 0
			for _, host := range handler.hosts {
				host.ping()
				if host.healthy {
					count ++
					available = true
				}
			}

			handler.healthyCount = count
			handler.available = available

			log.Printf("[%s] available: %t, healthy list: %d, up percent: %f", handler.path, handler.IsAvailable(), handler.getHealthyHosts(), handler.Status())

			time.Sleep(30 * time.Second)
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
