package lb

import (
	"github.com/coldog/proxy/lb/ctx"
	"github.com/coldog/proxy/lb/router"

	"errors"
	"net"
	"time"
	"github.com/coldog/proxy/lb/stats"
)

type Handler struct {
	Name                  string
	Routes                []*router.Route
	Strategy              string
	Middleware            []string
	Targets               []*Target
	MaxConn               int
	ShutdownWait          time.Duration
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	KeepAliveTimeout      time.Duration
	ReadTimeout           time.Duration
	DisableKeepAlives     bool
	DisableCompression    bool
	RawProxy              bool
	ClientIPHeader        string

	index         int
	currentWeight int

	quit      chan struct{}
	closed    bool
	draining  bool
	stats     stats.StatsCollector
}

func (h *Handler) Close() {
	close(h.quit)
	h.closed = true
}

func (h *Handler) AddTarget(t *Target) {
	h.Targets = append(h.Targets, t)
}

func (h *Handler) Process(c *ctx.Context) error {
	remoteIP, _, err := net.SplitHostPort(c.Req.RemoteAddr)
	if err != nil {
		return errors.New("cannot parse " + c.Req.RemoteAddr)
	}

	r := c.Req

	// set configurable ClientIPHeader
	// X-Real-Ip is set later and X-Forwarded-For is set
	// by the Go HTTP reverse proxy.
	if h.ClientIPHeader != "" && h.ClientIPHeader != "X-Forwarded-For" && h.ClientIPHeader != "X-Real-Ip" {
		r.Header.Set(h.ClientIPHeader, remoteIP)
	}

	if r.Header.Get("X-Real-Ip") == "" {
		r.Header.Set("X-Real-Ip", remoteIP)
	}

	// set the X-Forwarded-For header for websocket
	// connections since they aren't handled by the
	// http proxy which sets it.
	ws := r.Header.Get("Upgrade") == "websocket"
	if ws {
		r.Header.Set("X-Forwarded-For", remoteIP)
	}

	if r.Header.Get("X-Forwarded-Proto") == "" {
		switch {
		case ws && r.TLS != nil:
			r.Header.Set("X-Forwarded-Proto", "wss")
		case ws && r.TLS == nil:
			r.Header.Set("X-Forwarded-Proto", "ws")
		case r.TLS != nil:
			r.Header.Set("X-Forwarded-Proto", "https")
		default:
			r.Header.Set("X-Forwarded-Proto", "http")
		}
	}

	//if r.Header.Get("X-Forwarded-Port") == "" {
	//	r.Header.Set("X-Forwarded-Port", localPort(r))
	//}

	fwd := r.Header.Get("Forwarded")
	if fwd == "" {
		fwd = "for=" + remoteIP
		switch {
		case ws && r.TLS != nil:
			fwd += "; proto=wss"
		case ws && r.TLS == nil:
			fwd += "; proto=ws"
		case r.TLS != nil:
			fwd += "; proto=https"
		default:
			fwd += "; proto=http"
		}
	}

	//if h.LocalIP != "" {
	//	fwd += "; by=" + cfg.LocalIP
	//}
	//r.Header.Set("Forwarded", fwd)

	//if cfg.TLSHeader != "" && r.TLS != nil {
	//	r.Header.Set(cfg.TLSHeader, cfg.TLSHeaderValue)
	//}

	return nil
}

func (h *Handler) Next(c *ctx.Context) *Target {
	if h.closed || h.draining {
		return nil
	}

	dispatcher, ok := strategies[h.Strategy]
	if !ok {
		dispatcher = RRStrategy
	}

	h.stats.SetIncrement("requests." + h.Name, 1)
	return dispatcher(h, c)
}
