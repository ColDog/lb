package lb

import (
	"testing"
	"github.com/coldog/proxy/lb/router"
	"net/url"
	"net/http"
	"github.com/coldog/proxy/lb/ctx"
	"fmt"
)

func sample() *Handler {
	return &Handler{
		Name: "test",
		Routes: []*router.Route{
			{Path: "/tests/"},
			{Path: "/v1/api/*"},
		},
		Strategy: "rrh",
		Targets: []*Target{
			{
				ID: "test-1",
				URL: "http://localhost:3002",
				Weight: 10,
			},
			{
				ID: "test-2",
				URL: "http://localhost:3003",
				Weight: 20,
			},
		},
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

func TestStrategies_Divisor(t *testing.T) {
	h := sample()
	max, gcd := nums(h)
	if gcd != 4 && max != 12 {
		t.Fail()
	}
}

func TestStrategies_WRR(t *testing.T) {
	h := sample()

	for i := 0; i < 50; i++ {
		t := WRRStrategy(h, ctx.New(nil, mockReq("t", "t")))
		fmt.Println(t.ID)
	}
}
