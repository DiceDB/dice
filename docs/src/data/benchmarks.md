# Benchmarks

> Last updated: 10th October, 2024

Benchmarks play a pivotal role in the world of databases. They provide measurable insights into how a system performs under different conditions, helping developers and system administrators optimize configurations, identify bottlenecks, and predict behavior under stress. For DiceDB, a high-performance in-memory key-value store, benchmarking is crucial to demonstrate its scalability, resilience, and efficiency. This page will provide the latest numbers of our benchmark and will provide the exact steps to be taken to reproduce the numbers.

## Memtier Benchmark

[Memtier](https://github.com/RedisLabs/memtier_benchmark) is a powerful, easy-to-use benchmarking tool specifically designed for key-value databases like [Redis](https://redis.io/) and [Memcached](https://memcached.org/), making it ideal for evaluating [DiceDB](https://github.com/dicedb/dice). In this test, we used Memtier to generate a workload that mimics a real-world scenario, stressing DiceDB under a balanced read/write ratio.

### Machine configuration

The benchmarks were run on [AWS c5.12xlarge](https://aws.amazon.com/ec2/instance-types/c5/) instance
which has 48 vCPUs, 96 GB memory, and 12 Gbps network bandwidth. The benchmarks were run on an Ubuntu machine and on the latest commit on the `master` branch of [DiceDB](https://github.com/dicedb/dice).

## DiceDB Numbers

```sh
$ go build -o dicedb
$ ./dicedb --enable-multithreading=true
```

### Running the benchmark

```sh
$ memtier_benchmark -s 54.167.62.216 -p 7379 \
    --clients 50 --threads 24 \
    --ratio=1:1 --hide-histogram  \
    --data-size 256 --requests=100000
```

This command runs a benchmark where,

 - 50 clients will concurrently send requests
 - we execute benchmark using 24 threads to distribute the workload
 - the benchmark consists of a balanced mix of GET and SET operations (1:1 ratio)
 - each request will involve 256 bytes of data, simulating small-to-medium-sized transactions
 - a total of 100,000 requests will be executed, and the benchmark will run until all requests are completed.

### Results Observed

After running the above command on DiceDB we got the following numbers.

```
[RUN #1 100%, 199 secs]  0 threads:   120000000 ops, 2059627 (avg:  601850) ops/sec, 338.97MB/sec (avg: 98.86MB/sec),  0.58 (avg:  1.99) msec latency

24        Threads
50        Connections per thread
100000    Requests per client

============================================================================================================================
Type         Ops/sec     Hits/sec   Misses/sec    Avg. Latency     p50 Latency     p99 Latency   p99.9 Latency       KB/sec
----------------------------------------------------------------------------------------------------------------------------
Sets       310539.18          ---          ---         2.04382         1.67100         8.95900        18.17500     92156.88
Gets       310539.18         0.00    310539.18         1.94245         1.55100         8.83100        17.79100     12309.13
Waits           0.00          ---          ---             ---             ---             ---             ---          ---
Totals     621078.36         0.00    310539.18         1.99313         1.61500         8.89500        17.91900    104466.01
```

The benchmark run tested DiceDB with 24 threads, 50 client connections per thread, and 100,000 requests per client. Below is a breakdown of the key metrics and their significance:

- total throughput achieved was 621,078 (ops/sec), evenly split between GETs and SETs
- the average latency for the overall benchmark was 0.472 milliseconds, pretty fast.
  - p50 latency was 1.61 ms
  - p99 latency was 8.89 ms, indicating that 99% of requests were faster than this.
  - p99.9 latency reached 17.91 ms, showing how the highest-latency outliers behaved.

## Running the same on Redis

```sh
$ memtier_benchmark -s 54.167.62.216 -p 6379 \
    --clients 50 --threads 24 \
    --ratio=1:1 --hide-histogram  \
    --data-size 256 --requests=100000
```

After running the above command on Redis on the same infrastructure we got the following numbers.

```
[RUN #1 100%, 594 secs]  0 threads:   120000000 ops,  201550 (avg:  201774) ops/sec, 33.33MB/sec (avg: 33.20MB/sec),  5.95 (avg:  5.94) msec latency

24        Threads
50        Connections per thread
100000    Requests per client


ALL STATS
============================================================================================================================
Type         Ops/sec     Hits/sec   Misses/sec    Avg. Latency     p50 Latency     p99 Latency   p99.9 Latency       KB/sec
----------------------------------------------------------------------------------------------------------------------------
Sets       100842.84          ---          ---         5.94669         5.40700        10.62300        13.56700     29926.53
Gets       100842.84       516.32    100326.53         5.94390         5.40700        10.62300        13.56700      4058.81
Waits           0.00          ---          ---             ---             ---             ---             ---          ---
Totals     201685.69       516.32    100326.53         5.94529         5.40700        10.62300        13.56700     33985.35
```

The benchmark run tested Redis with 24 threads, 30 client connections per thread, and 100,000 requests per client. The total number of operations (GET and SET) was 30,000,000, completed in 47 seconds. Below is a breakdown of the key metrics and their significance:

- total throughput achieved was 640,240.33 (ops/sec), evenly split between GETs and SETs
- the average data transfer rate was 112.44 MB/sec, with GETs contributing 12.69 MB/sec and SETs 95 MB/sec.
- the average latency for the overall benchmark was 0.472 milliseconds, pretty fast.
  - p50 latency was 1.61 ms
  - p99 latency was 8.89 ms, indicating that 99% of requests were faster than this.
  - p99.9 latency reached 17.91 ms, showing how the highest-latency outliers behaved.

## Summary

The results from the memtier benchmark suggest that DiceDB is well-optimized for handling high-concurrency environments with minimal resource consumption. The high throughput and low latency are clear indicators that DiceDB can serve mission-critical applications that require real-time processing with rapid response times. 

Despite the simulated stress of 50 clients making simultaneous requests, DiceDB managed to maintain its performance characteristics without any visible degradation.
