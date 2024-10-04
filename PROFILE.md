```sh
in tab 1

$ go build -gcflags="all=-N -l" -o dicedb main.go
$ ./dicedb --enable-multithreading=true

in tab 2

$ docker run --network host redislabs/memtier_benchmark memtier_benchmark -s host.docker.internal  -p 7379

let benchmark complete in tab 2
stop the dicedb server in tab 1 and let it to flush the cpu.pprof file

$ go tool pprof -http=:8080 cpu.pprof 
```
