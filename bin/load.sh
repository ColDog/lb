#!/bin/bash

node bin/hello_world.js 8001
node bin/hello_world.js 8002
node bin/hello_world.js 8003
node bin/hello_world.js 8004

wrk http://localhost:3000


