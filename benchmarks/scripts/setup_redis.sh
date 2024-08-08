#!/bin/bash

git clone https://github.com/redis/redis/
cd redis
git checkout 7.2.5
make
cd src
nohup ./redis-server > /dev/null 2>&1 &
