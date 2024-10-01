# Benchmarks

> Last updated: 1st October, 2024

Benchmarks play a pivotal role in the world of databases. They provide measurable insights into how a system performs under different conditions, helping developers and system administrators optimize configurations, identify bottlenecks, and predict behavior under stress. For DiceDB, a high-performance in-memory key-value store, benchmarking is crucial to demonstrate its scalability, resilience, and efficiency. This page will provide the latest numbers of our benchmark and will provide the exact steps to be taken to reproduce the numbers.

## Memtier Benchmark

[Memtier](https://github.com/RedisLabs/memtier_benchmark) is a powerful, easy-to-use benchmarking tool specifically designed for key-value databases like [Redis](https://redis.io/) and [Memcached](https://memcached.org/), making it ideal for evaluating [DiceDB](https://github.com/dicedb/dice). In this test, we used Memtier to generate a workload that mimics a real-world scenario, stressing DiceDB under a balanced read/write ratio.

### Machine configuration

The benchmarks were run on [AWS c5.12xlarge](https://aws.amazon.com/ec2/instance-types/c5/) instance
which has 48 vCPUs, 96 GB memory, and 12 Gbps network bandwidth. The benchmarks were run on an Ubuntu machine and on the latest commit on the `master` branch of [DiceDB](https://github.com/dicedb/dice).

### Run DiceDB

```sh
$ go build -o dicedb
$ ./dicedb --enable-multithreading=true
```

### Running the benchmark

```sh
$ memtier_benchmark -s localhost -p 7379 \
    --clients 30 --threads 10 \
    --ratio=1:1 --hide-histogram \
    --data-size 256 --requests=100000
```

This command runs a benchmark where,

 - 30 clients will concurrently send requests
 - we execute benchmark using 10 threads to distribute the workload
 - the benchmark consists of a balanced mix of GET and SET operations (1:1 ratio)
 - each request will involve 256 bytes of data, simulating small-to-medium-sized transactions
 - a total of 100,000 requests will be executed, and the benchmark will run until all requests are completed.

### Results Observed

After running the above command on DiceDB we got the following numbers.

```
============================================================================================================================
Type         Ops/sec     Hits/sec   Misses/sec    Avg. Latency     p50 Latency     p99 Latency   p99.9 Latency       KB/sec
----------------------------------------------------------------------------------------------------------------------------
Sets       320120.17          ---          ---         0.50704         0.40700         2.01500         4.63900     95000.17
Gets       320120.17       857.26    319262.91         0.43723         0.31100         1.94300         4.54300     12686.75
Waits           0.00          ---          ---             ---             ---             ---             ---          ---
Totals     640240.33       857.26    319262.91         0.47214         0.35900         1.98300         4.60700    107686.92
```

The benchmark run tested DiceDB with 10 threads, 30 client connections per thread, and 100,000 requests per client. The total number of operations (GET and SET) was 30,000,000, completed in 47 seconds. Below is a breakdown of the key metrics and their significance:

- total throughput achieved was 640,240.33 (ops/sec), evenly split between GETs and SETs
- the average data transfer rate was 112.44 MB/sec, with GETs contributing 12.69 MB/sec and SETs 95 MB/sec.
- the average latency for the overall benchmark was 0.472 milliseconds, pretty fast.
  - p50 latency was 0.359 ms
  - p99 latency was 1.983 ms, indicating that 99% of requests were faster than this.
  - p99.9 latency (99.9th percentile) reached 4.607 ms, showing how the highest-latency outliers behaved.

The results from the memtier benchmark suggest that DiceDB is well-optimized for handling high-concurrency environments with minimal resource consumption. The high throughput and low latency are clear indicators that DiceDB can serve mission-critical applications that require real-time processing with rapid response times. 

Despite the simulated stress of 30 clients making simultaneous requests, DiceDB managed to maintain its performance characteristics without any visible degradation.
