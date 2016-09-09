package lb

import (
	"github.com/coldog/proxy/lb/stats"

	"golang.org/x/net/websocket"

	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Target struct {
	ID             string
	URL            string
	Timeout        int
	Weight         int
	FailedRequests int

	proxy          http.Handler
	rawProxy       http.Handler
	wsProxy        http.Handler
	url            *url.URL
	tr             *http.Transport
	stats          stats.StatsCollector
}

func (b *Target) Proxy(h *Handler, r *http.Request) http.Handler {
	if b.url == nil {
		u, err := url.Parse(b.URL)
		if err != nil {
			return nil
		}

		b.url = u
	}

	if b.tr == nil {
		b.tr = &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   h.DialTimeout,
				KeepAlive: h.KeepAliveTimeout,
				Cancel:    h.quit,
			}).Dial,
			DisableKeepAlives:     h.DisableKeepAlives,
			ResponseHeaderTimeout: h.ResponseHeaderTimeout,
			ExpectContinueTimeout: h.DialTimeout,
			MaxIdleConnsPerHost:   h.MaxConn,
			DisableCompression:    h.DisableCompression,
		}
	}

	if h.RawProxy {
		if b.rawProxy == nil {
			b.rawProxy = newRawProxy(b.url)
		}
		return b.rawProxy

	} else {
		if r.Header.Get("Upgrade") == "websocket" {
			if b.wsProxy == nil {
				b.wsProxy = newWSProxy(b.url)
			}
			return b.wsProxy
		} else {
			if b.proxy == nil {
				b.proxy = newHTTPProxyWithTripper(b, time.Duration(0))
			}
			return b.proxy
		}
	}
}

func newHTTPProxyWithTripper(t *Target, flush time.Duration) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t.url)
	rp.FlushInterval = flush
	rp.Transport = &meteredRoundTripper{
		id: t.ID,
		tr: t.tr,
		stat: t.stats,
		t: t,
	}
	return rp
}

type meteredRoundTripper struct {
	id string
	tr http.RoundTripper
	stat stats.StatsCollector
	t *Target
}

func (m *meteredRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	t1 := time.Now()
	resp, err := m.tr.RoundTrip(r)


	m.stat.SetTime(m.id, t1)
	m.stat.SetIncrement(m.id + "." + statusCodeName(resp), 1)
	if err != nil || resp.StatusCode >= 500 {
		m.t.FailedRequests += 1
	}
	return resp, err
}

func newRawProxy(t *url.URL) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "not a hijacker", http.StatusInternalServerError)
			return
		}

		in, _, err := hj.Hijack()
		if err != nil {
			log.Printf("[ERROR] Hijack error for %s. %s", r.URL, err)
			http.Error(w, "hijack error", http.StatusInternalServerError)
			return
		}
		defer in.Close()

		out, err := net.Dial("tcp", t.Host)
		if err != nil {
			log.Printf("[ERROR] WS error for %s. %s", r.URL, err)
			http.Error(w, "error contacting backend server", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		err = r.Write(out)
		if err != nil {
			log.Printf("[ERROR] Error copying request for %s. %s", r.URL, err)
			http.Error(w, "error copying request", http.StatusInternalServerError)
			return
		}

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}

		go cp(out, in)
		go cp(in, out)
		err = <-errc
		if err != nil && err != io.EOF {
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
		}
	})
}

func newWSProxy(t *url.URL) http.Handler {
	return websocket.Handler(func(in *websocket.Conn) {
		defer in.Close()

		r := in.Request()
		targetURL := "ws://" + t.Host + r.RequestURI
		out, err := websocket.Dial(targetURL, "", r.Header.Get("Origin"))
		if err != nil {
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
			return
		}
		defer out.Close()

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}

		go cp(out, in)
		go cp(in, out)
		err = <-errc
		if err != nil && err != io.EOF {
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
		}
	})
}

func statusCodeName(r *http.Response) string {
	if r == nil {
		return "5xx"
	}

	code := r.StatusCode

	if code >= 500 {
		return "5xx"
	} else if code < 500 && code >= 400 {
		return "4xx"
	} else if code < 400 && code >= 300 {
		return "3xx"
	} else if code < 300 && code >= 200 {
		return "2xx"
	} else {
		return "xxx"
	}
}