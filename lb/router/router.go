package router

import (
	"regexp"
	"net/http"
)

type Route struct {
	Path     string
	Host     string
	Priority int
	key      string
	pathRegx *regexp.Regexp
	hostRegx *regexp.Regexp
	headRegx *regexp.Regexp
}

func New() *Router {
	return &Router{[]*Route{}}
}

type Router struct {
	routes []*Route
}

func (r *Router) Remove(key string) {
	for i, route := range r.routes {
		if route.key == key {
			r.routes = append(r.routes[0:i], r.routes[i+1:]...)
		}
	}
}

func (r *Router) Add(key string, route *Route) (err error) {
	route.key = key

	if route.Path != "" {
		route.pathRegx, err = regexp.Compile(route.Path)
		if err != nil {
			return err
		}
	}

	if route.Host != "" {
		route.hostRegx, err = regexp.Compile(route.Host)
		if err != nil {
			return err
		}
	}

	if len(r.routes) > 0 {
		if route.Priority >= r.routes[0].Priority {
			r.routes = append([]*Route{route}, r.routes...)
		} else if route.Priority <= r.routes[len(r.routes) - 1].Priority {
			r.routes = append(r.routes, route)
		} else {
			for i := 0; i < len(r.routes)-1; i++ {
				if route.Priority <= r.routes[i].Priority && route.Priority >= r.routes[i+1].Priority {
					r.routes = append(r.routes[0:i], append([]*Route{route}, r.routes[i:]...)...)
					break
				}
			}
		}
	} else {
		r.routes = append(r.routes, route)
	}

	return nil
}

func (r *Router) Match(req *http.Request) string {
	for _, route := range r.routes {
		if route.pathRegx != nil && route.pathRegx.MatchString(req.URL.Path) {
			return route.key
		}

		if route.hostRegx != nil && route.hostRegx.MatchString(req.Host) {
			return route.key
		}
	}
	return ""
}
