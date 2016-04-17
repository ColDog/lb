##### Benchmarks
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

