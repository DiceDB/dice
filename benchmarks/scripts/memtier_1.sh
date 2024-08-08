#!/bin/bash
set -e

memtier_benchmark -s localhost -p 6379 -c 50 -t 10 \
    --key-pattern="R:R" --data-size=64 --ratio=1:1 --out-file=/home/ubuntu/results-1.txt
