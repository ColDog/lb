# Golang Load Balancer

This is a dynamically configured load balancer. You provide it with a list of routes and hosts, and it will dynamically balance http traffic between the hosts.

### Configuration
```
Example configuration:
{
"key": "test",
"ip_hash": false,
"routes": [
		"/test",
		"/test/*"
	],
"middleware": [
		"JwtAuth"
	],
"hosts": [
    {
			"target": "http://localhost:8000",
			"health": "http://localhost:8000/health",
			"timeout": 10,
			"down": false
    }
	]
}
```

### Benchmarks
```
Setup: 7 hello world node.js servers

$  wrk http://localhost:3000/users/`
>  Running 10s test @ http://localhost:3000/users/
>    2 threads and 10 connections
>    Thread Stats   Avg      Stdev     Max   +/- Stdev
>      Latency   602.25us  519.56us  13.94ms   92.96%
>      Req/Sec     6.15k     2.19k   11.21k    52.97%
>    123764 requests in 10.10s, 15.93MB read
>  Requests/sec:  12250.95
>  Transfer/sec:      1.58MB
```

