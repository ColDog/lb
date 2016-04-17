#!/usr/local/bin/node

var http = require('http');
var port = parseInt(process.argv[2]);

var server = http.createServer(function (request, response) {
    console.log("request from", request.url);
    response.writeHead(200, {"Content-Type": "application/json"});
    response.end('{"msg": "Hello World", "port": "' + port + '"}\n');
});


server.listen(port);

console.log("Server running at http://127.0.0.1:" + port);
