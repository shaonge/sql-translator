#!/usr/bin/env bash

docker pull mysql:5.7

docker run --name mysql -p 3306:3306 -e MYSQL\_ROOT\_PASSWORD=123456 -d mysql