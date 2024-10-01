## Pipelining Example

```
PING:       *1\r\n$4\r\nPING\r\n
SET k v:    *3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n
GET k:      *2\r\n$3\r\nGET\r\n$1\r\nk\r\n
```

```
$ (printf 'CMD1CMD2CMD3';) | nc localhost 7379
$ (printf '*1\r\n$4\r\nPING\r\n*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n*2\r\n$3\r\nGET\r\n$1\r\nk\r\n';) | nc localhost 7379
```

## Storm

Storm is a series of utility that allows us to fire bulk request to the Dice Database

```
$ go run storm/set/set.go
```

## Monitoring through Prometheus

```
$ ./redis_exporter  -redis.addr redis://localhost:7379
$ ./prometheus --web.enable-admin-api
$ curl -X POST -g 'http://localhost:9090/api/v1/admin/tsdb/delete_series?match[]={instance="localhost:9121"}'
```
