package main

import (
	"net/http"
	"time"
	"log"
)

type Host struct {
	target		string
	health		string
	proxy 		http.Handler
	checked		int64
	healthy		bool
}

func (host Host) ping() {
	resp, err := http.Get(host.health)
	host.healthy = err == nil && resp.StatusCode == 200
	host.checked = time.Now().Unix()
	log.Printf("pinging %s success: %s\n", host.target, host.healthy)
}

type Handler struct {
	path 		string
	hosts		[]Host
	middleware	[]string
	lastHost	int
}

func (handler *Handler) next() Host {
	if len(handler.hosts) == 1 {
		return handler.hosts[0]

	} else if len(handler.hosts) > 1 {
		handler.lastHost++

		if handler.lastHost >= len(handler.hosts) {
			handler.lastHost = 0
		}

		for ; handler.lastHost < len(handler.hosts); handler.lastHost++ {
			if handler.hosts[handler.lastHost].healthy {
				break
			}
		}

		return handler.hosts[handler.lastHost]
	}

	panic("Cannot find host to handle request")
}

func (handler *Handler) run(host *Host, writer http.ResponseWriter, request *http.Request) (http.ResponseWriter, *http.Request) {
	req := &Request{handler, request, host}
	res := &Response{writer}

	for _, middleKey := range handler.middleware {
		middleware[middleKey](req, res)
	}

	return res.writer, req.req
}
