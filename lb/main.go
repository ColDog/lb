package main

import (
	"github.com/coldog/proxy/lb/lb"
	"github.com/coldog/proxy/lb/router"
)

func main()  {
	l := lb.New(lb.DefaultConfig())

	l.PutHandler(&lb.Handler{
		Name: "test",
		Routes: []*router.Route{
			{Path: "/tests/"},
		},
		RawProxy: true,
		Strategy: "ip_hash",
		Targets: []*lb.Target{
			{
				ID: "test-1",
				URL: "http://localhost:3000",
				Weight: 10,
			},
			{
				ID: "test-2",
				URL: "http://localhost:3001",
				Weight: 15,
			},
			{
				ID: "test-3",
				URL: "http://localhost:3002",
				Weight: 15,
			},
			{
				ID: "test-4",
				URL: "http://localhost:3003",
				Weight: 15,
			},
		},
	})

	l.Start()
}
