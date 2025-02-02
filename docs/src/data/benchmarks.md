# Benchmarks on Performance and Scalability

> Last updated: 4th November, 2024

Benchmarks play a pivotal role in the world of databases. They provide measurable insights into how a system performs under different conditions, helping developers optimize configurations, identify bottlenecks, and predict behavior under stress.

For DiceDB, a high-performance in-memory key-value store, benchmarking is crucial to demonstrate its scalability, resilience, and efficiency. Here's our latest benchmark that will demonstrate our numbers, compare them with Redis, and more importantly provide the exact steps to be taken to reproduce the numbers.

Redis, created over a decade ago, relies on a basic single-threaded architecture, which can make scaling deployments a challenging and time-consuming task for developers. In contrast, DiceDB offers a seamless alternative to Redis. As a modern in-memory database, DiceDB builds on the strengths of earlier systems but enhances both performance and reliability.

With its multi-threaded design, DiceDB supports efficient vertical scaling, allowing you to sidestep the complexities of horizontal scaling and cluster management, all while achieving the scalability your application requires. Because DiceDB is a drop-in replacement of Redis, we can run the standard [Memtier](https://github.com/RedisLabs/memtier_benchmark) benchmark to compare the two.

## Memtier Benchmark

[Memtier](https://github.com/RedisLabs/memtier_benchmark) is a powerful, easy-to-use benchmarking tool specifically designed for key-value databases like [Redis](https://redis.io/) and [Memcached](https://memcached.org/), making it ideal for evaluating [DiceDB](https://github.com/dicedb/dice). In this test, we used Memtier to generate a workload that mimics a real-world scenario, stressing DiceDB under a balanced read/write ratio.

### Machine configuration

The benchmarks were run on [AWS c5.12xlarge](https://aws.amazon.com/ec2/instance-types/c5/) instance
which has 48 vCPUs, 96 GB memory, and 12 Gbps network bandwidth. The benchmarks were run on an Ubuntu machine and the latest commit was on the `master` branch of [DiceDB](https://github.com/dicedb/dice).

### Infrastructure Setup

Take two AWS c5.12xlarge machines, one machine to run memtier while the other to run DiceDB.

## DiceDB Benchmark

On the machine that is designated to run DiceDB, fire the following commands to build and run DiceDB.

```sh
$ go build -o dicedb
$ ./dicedb
```

### Running the benchmark

On the machine that is designated to run memtier, fire the following commands to start the benchmark.

```sh
$ memtier_benchmark -s 54.204.230.210 -p 7379 \
 --clients 50 --threads 24 \
    --ratio=1:1 --hide-histogram \
 --data-size 256 --requests=100000
```

This command runs a benchmark where,

- we execute benchmark using 24 threads to distribute the workload
- each thread will set 50 clients and send requests concurrently
- the benchmark consists of a balanced mix of GET and SET operations (1:1 ratio)
- each request will involve 256 bytes of data, simulating small-to-medium-sized transactions
- a total of 100,000 requests will be executed, and the benchmark will run until all requests are completed.

### Results Observed

After running the above command on DiceDB we got the following numbers.

```
[RUN #1] Preparing benchmark client...
[RUN #1] Launching threads now...
[RUN #1 100%, 223 secs]  0 threads:   120000000 ops, 1940109 (avg:  536901) ops/sec, 319.22MB/sec (avg: 88.20MB/sec),  0.62 (avg:  2.23) msec latency

24        Threads
50        Connections per thread
100000    Requests per client


ALL STATS
============================================================================================================================
Type         Ops/sec     Hits/sec   Misses/sec    Avg. Latency     p50 Latency     p99 Latency   p99.9 Latency       KB/sec
----------------------------------------------------------------------------------------------------------------------------
Sets       275447.56          ---          ---         2.26364         1.35100         8.95900        13.63100     81742.94
Gets       275447.56         0.00    275447.56         2.20482         1.31900         8.76700        13.37500     10925.92
Waits           0.00          ---          ---             ---             ---             ---             ---          ---
Totals     550895.11         0.00    275447.56         2.23423         1.33500         8.83100        13.50300     92668.87
```

The benchmark run tested DiceDB with 24 threads, 50 client connections per thread, and 100,000 requests per client which turns out to be 120,000,000 operations in total. Below is a breakdown of the key metrics and their significance:

- the peak throughput DiceDB can handle is 1,940,109 ops/sec
- the average throughput was 536,901 ops/sec, evenly split between GETs and SETs
- the average latency for the overall benchmark was 1.99 milliseconds, pretty fast.
    - p50 latency was 1.33 ms
    - p99 latency was 8.83 ms, indicating that 99% of requests were faster than this.
    - p99.9 latency reached 13.50 ms, showing how the highest-latency outliers behaved.

## Redis Benchmark

Now, you can stop the DiceDB process setup Redis, and run the server on port 6379.

```sh
$ ./redis-server
```

### Running the benchmark

```sh
$ memtier_benchmark -s 3.82.51.57 -p 6379 \
 --clients 50 --threads 24 \
    --ratio=1:1 --hide-histogram \
 --data-size 256 --requests=100000
```

### Results Observed

After running the above command on Redis we got the following numbers.

```
[RUN #1] Preparing benchmark client...
[RUN #1] Launching threads now...
[RUN #1 100%, 623 secs]  0 threads:   120000000 ops,  202925 (avg:  192485) ops/sec, 33.44MB/sec (avg: 31.67MB/sec),  5.91 (avg:  6.23) msec latency

24        Threads
50        Connections per thread
100000    Requests per client


ALL STATS
============================================================================================================================
Type         Ops/sec     Hits/sec   Misses/sec    Avg. Latency     p50 Latency     p99 Latency   p99.9 Latency       KB/sec
----------------------------------------------------------------------------------------------------------------------------
Sets        96224.94          ---          ---         6.23382         5.69500        11.26300        14.01500     28556.11
Gets        96224.94       492.67     95732.27         6.23054         5.69500        11.26300        13.95100      3872.95
Waits           0.00          ---          ---             ---             ---             ---             ---          ---
Totals     192449.89       492.67     95732.27         6.23218         5.69500        11.26300        14.01500     32429.05
```

The benchmark run tested Redis with 24 threads, 50 client connections per thread, and 100,000 requests per client which turns out to be 120,000,000 operations in total. Below is a breakdown of the key metrics and their significance:

- the peak throughput Redis can handle is 202,925 ops/sec
- the average throughput was 192,485 ops/sec, evenly split between GETs and SETs
- the average latency for the overall benchmark was 1.99 milliseconds, pretty fast.
    - p50 latency was 5.69 ms
    - p99 latency was 11.26 ms, indicating that 99% of requests were faster than this.
    - p99.9 latency reached 14.01 ms, showing how the highest-latency outliers behaved.

## DiceDB vs Redis

| Metric                       | DiceDB                            | Redis                             |
| ---------------------------- | --------------------------------- | --------------------------------- |
| Peak Throughput (ops/sec)    | 1,940,109                         | 202,925                           |
| Average Throughput (ops/sec) | 536,901                           | 192,485                           |
| Average Latency (ms)         | 2.23                              | 6.23                              |
| p50 Latency (ms)             | 1.33                              | 5.69                              |
| p99 Latency (ms)             | 8.83                              | 11.26                             |
| p99.9 Latency (ms)           | 13.50                             | 14.01                             |

The benchmark results demonstrate that DiceDB significantly outperforms Redis in terms of throughput. DiceDB's peak throughput is ~9 times higher than Redis, reaching **1,940,109 ops per second** compared to Redis' 202,925 ops/sec. Additionally, DiceDB's average throughput is also much higher, clocking in at **536,901 ops/sec**, while Redis achieves 192,485 ops/sec. This indicates that DiceDB can handle far more operations in a given period, making it a better choice for applications requiring high throughput, such as real-time or reactive systems.

In terms of latency, DiceDB also offers better average latency at **2.23 milliseconds**. However, when analyzing the different latency percentiles, DiceDB performs better across the spectrum - average, p50, p99, and p99.9 latencies.

### DiceDB vs Redis - number of cores

If we run the above benchmark while altering the number of cores of the underlying hardware, we can see how DiceDB is optimized for scalability. Redis exhibits stable throughput across core counts, suggesting a limit to its ability to leverage additional processing power. By contrast, DiceDB's throughput increases as more cores are added, indicating that it dynamically scales with available hardware resources.

This difference showcases DiceDB's design advantage in multi-core environments, where it effectively leverages all the underlying cores to offer maximum performance. This not only highlights DiceDB's potential for high-throughput applications but also its efficiency in using hardware resources to reduce latency and improve responsiveness.

![DiceDB vs Redis - Throughput vs Number of Cores](https://github.com/user-attachments/assets/367f25c8-bc59-4787-bb37-63fd5d4314e4)

Overall, DiceDB offers superior throughput and generally lower latency, making it a more robust solution for high-performance, low-latency applications, especially in cases where extreme latency outliers are less critical.
