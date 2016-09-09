package router

import (
	"testing"
	"net/http"
	"net/url"
)

var r *Router

func init()  {
	r = New()

	r.Add("t1", &Route{
		Host: "api.stuff.*",
		Priority: 5,
	})

	r.Add("t2", &Route{
		Path: "/api/*",
		Priority: 10,
	})

	r.Add("t3", &Route{
		Path: "/v2/api/*",
		Priority: 7,
	})

	r.Add("t4", &Route{
		Path: "/v3/api/*",
		Priority: 7,
	})

	r.Add("t5", &Route{
		Path: "/v4/api/*",
		Priority: 7,
	})

	for i := 0; i < 10000; i++ {
		r.Add("t5", &Route{
			Path: "/v4/api/*",
			Priority: 20,
		})
	}
}

func mockReq(host, path string) *http.Request {
	u, err := url.Parse("http://" + host + "/" + path)
	if err != nil {
		panic(err)
	}
	return &http.Request{
		Host: host,
		Method: "GET",
		URL: u,
	}
}

func TestRouting(t *testing.T) {

	r := New()

	r.Add("t1", &Route{
		Host: "api.stuff.*",
		Priority: 5,
	})

	r.Add("t2", &Route{
		Path: "/api/*",
		Priority: 10,
	})

	r.Add("t3", &Route{
		Path: "/v2/api/*",
		Priority: 7,
	})

	m := r.Match(mockReq("api.stuff.com", "stuff"))
	if m != "t1" {
		t.Fail()
	}

	m = r.Match(mockReq("api.stuff.com", "/api/stuff"))
	if m != "t2" {
		t.Fail()
	}

	m = r.Match(mockReq("api.stuff.com", "/v2/api/stuff"))
	if m != "t3" {
		t.Fail()
	}
}

var result string
func BenchmarkRouter(b *testing.B) {
	var m string
	req := mockReq("api.stuff.com", "/v2/api/stuff")

	for n := 0; n < b.N; n++ {
		m = r.Match(req)
	}

	result = m
}
