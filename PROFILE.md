```sh
$ docker run --network host redislabs/memtier_benchmark memtier_benchmark -s host.docker.internal  -p 7379
$ stop the dicedb server to flush the cpu.pprof file

$ go tool pprof -http=:8080 cpu.pprof 
```
