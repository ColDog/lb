#!/usr/local/bin/node

var http = require('http');

var server = http.createServer(function (request, response) {
    console.log("request from", request.url);
    response.writeHead(200, {"Content-Type": "application/json"});
    response.end('{"msg": "Hello World"}\n');
});

var port = parseInt(process.argv[2]);

server.listen(port);

console.log("Server running at http://127.0.0.1:" + port);
