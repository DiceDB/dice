# Benchmarks on Performance and Scalability

> Last updated: 10th October, 2024

Benchmarks play a pivotal role in the world of databases. They provide measurable insights into how a system performs under different conditions, helping developers and system administrators optimize configurations, identify bottlenecks, and predict behavior under stress. For DiceDB, a high-performance in-memory key-value store, benchmarking is crucial to demonstrate its scalability, resilience, and efficiency. This article will provide the latest numbers of our benchmark and will provide the exact steps to be taken to reproduce the numbers.

Redis, created over a decade ago, relies on a basic single-threaded architecture, which can make scaling deployments a challenging and time-consuming task for developers. In contrast, DiceDB offers a seamless alternative to Redis. As a modern in-memory database, DiceDB builds on the strengths of earlier systems but enhances both performance and reliability. With its multi-threaded design, DiceDB supports efficient vertical scaling, allowing you to sidestep the complexities of horizontal scaling and cluster management, all while achieving the scalability your application requires.

Because DiceDB is a drop-in replacement of Redis, we can run the standard [Memtier](https://github.com/RedisLabs/memtier_benchmark) benchmark to compare the two.

## Memtier Benchmark

[Memtier](https://github.com/RedisLabs/memtier_benchmark) is a powerful, easy-to-use benchmarking tool specifically designed for key-value databases like [Redis](https://redis.io/) and [Memcached](https://memcached.org/), making it ideal for evaluating [DiceDB](https://github.com/dicedb/dice). In this test, we used Memtier to generate a workload that mimics a real-world scenario, stressing DiceDB under a balanced read/write ratio.

### Machine configuration

The benchmarks were run on [AWS c5.12xlarge](https://aws.amazon.com/ec2/instance-types/c5/) instance
which has 48 vCPUs, 96 GB memory, and 12 Gbps network bandwidth. The benchmarks were run on an Ubuntu machine and on the latest commit on the `master` branch of [DiceDB](https://github.com/dicedb/dice).

## DiceDB Benchmark

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

- we execute benchmark using 24 threads to distribute the workload
- each thread will setup 50 clients and send requests concurrently
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

The benchmark run tested DiceDB with 24 threads, 50 client connections per thread, and 100,000 requests per client which turns out to be 120,000,000 operations in total. Below is a breakdown of the key metrics and their significance:

- the peak throughput DiceDB can handle is 2,059,627 ops/sec
- the average throughput was 621,078 ops/sec, evenly split between GETs and SETs
- the average latency for the overall benchmark was 1.99 milliseconds, pretty fast.
  - p50 latency was 1.61 ms
  - p99 latency was 8.89 ms, indicating that 99% of requests were faster than this.
  - p99.9 latency reached 17.91 ms, showing how the highest-latency outliers behaved.

## Redis Benchmark

```sh
$ go build -o dicedb
$ ./dicedb --enable-multithreading=true
```

### Running the benchmark

```sh
$ memtier_benchmark -s 54.167.62.216 -p 6379 \
    --clients 50 --threads 24 \
    --ratio=1:1 --hide-histogram  \
    --data-size 256 --requests=100000
```

This command runs a benchmark where,

- we execute benchmark using 24 threads to distribute the workload
- each thread will setup 50 clients and send requests concurrently
- the benchmark consists of a balanced mix of GET and SET operations (1:1 ratio)
- each request will involve 256 bytes of data, simulating small-to-medium-sized transactions
- a total of 100,000 requests will be executed, and the benchmark will run until all requests are completed.

### Results Observed

After running the above command on Redis we got the following numbers.

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

The benchmark run tested Redis with 24 threads, 50 client connections per thread, and 100,000 requests per client which turns out to be 120,000,000 operations in total. Below is a breakdown of the key metrics and their significance:

- the peak throughput Redis can handle is 201,550 ops/sec
- the average throughput was 201,774 ops/sec, evenly split between GETs and SETs
- the average latency for the overall benchmark was 1.99 milliseconds, pretty fast.
  - p50 latency was 5.94 ms
  - p99 latency was 10.62 ms, indicating that 99% of requests were faster than this.
  - p99.9 latency reached 13.56 ms, showing how the highest-latency outliers behaved.

## DiceDB vs Redis

| Metric                | DiceDB                       | Redis                        |
|---------------------------|-----------------------------------|-----------------------------------|
| Peak Throughput (ops/sec) | 2,059,627                     | 201,550                          |
| Average Throughput (ops/sec) | 621,078                    | 201,774                          |
| Average Latency (ms)   | 1.99                             | 1.99                             |
| p50 Latency (ms)       | 1.61                             | 5.94                             |
| p99 Latency (ms)       | 8.89                             | 10.62                            |
| p99.9 Latency (ms)     | 17.91                            | 13.56                            |

The benchmark results clearly demonstrate that DiceDB significantly outperforms Redis in terms of throughput. DiceDB's peak throughput is over 10 times higher than Redis, reaching **2,059,627 ops per second** compared to Redis' 201,550 ops/sec. Additionally, DiceDB's average throughput is also much higher, clocking in at **621,078 ops/sec**, while Redis achieves 201,774 ops/sec. This indicates that DiceDB can handle far more operations in a given period, making it a better choice for applications requiring high throughput, such as real-time or reactive systems.

In terms of latency, both databases show similar average latency at **1.99 milliseconds**. However, when analyzing the different latency percentiles, DiceDB performs better at p50 and p99 latencies, with faster response times than Redis, particularly at p50, where DiceDB achieves **1.61 ms** compared to Redis’ 5.94 ms. Interestingly, DiceDB has a higher p99.9 latency at 17.91 ms, indicating that although most requests are faster, it has some outliers with higher latency than Redis.

Overall, DiceDB offers superior throughput and generally lower latency, making it a more robust solution for high-performance, low-latency applications, especially in cases where extreme latency outliers are less critical.
