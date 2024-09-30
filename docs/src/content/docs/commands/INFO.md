---
title: INFO
description: The `INFO` command in DiceDB is used to retrieve information and statistics about the DiceDB server. This command provides a wealth of information about the server's status, including memory usage, CPU usage, keyspace statistics, and more. The information is returned in a format that is easy to parse and understand.
---

The `INFO` command in DiceDB is used to retrieve information and statistics about the DiceDB server. This command provides a wealth of information about the server's status, including memory usage, CPU usage, keyspace statistics, and more. The information is returned in a format that is easy to parse and understand.

## Parameters

### Optional Section Parameter

- `section`: (Optional) A string specifying the section of information to retrieve. If no section is specified, the command returns all sections. The available sections are:
- `server`: General information about the DiceDB server.
- `clients`: Client connections information.
- `memory`: Memory usage information.
- `persistence`: Information about RDB and AOF persistence.
- `stats`: General statistics.
- `replication`: Master/slave replication information.
- `cpu`: CPU usage statistics.
- `commandstats`: Command statistics.
- `cluster`: DiceDB Cluster information.
- `keyspace`: Database keyspace statistics.
- `modules`: Information about loaded modules.

## Return Value

The `INFO` command returns a bulk string containing the requested information. The information is formatted as a series of lines, each containing a key-value pair separated by a colon. Each section is separated by a blank line and starts with a line containing the section name in square brackets.

## Example Usage

### Retrieve All Information

```bash
127.0.0.1:7379> INFO

DiceDB_version:6.2.5
DiceDB_git_sha1:00000000
DiceDB_git_dirty:0
DiceDB_build_id:bf7a1e1b1e1b1e1b
DiceDB_mode:standalone
os:Linux 4.15.0-112-generic x86_64
arch_bits:64
multiplexing_api:epoll
gcc_version:7.5.0
process_id:1
run_id:bf7a1e1b1e1b1e1b1e1b1e1b1e1b1e1b1e1b1e1b
tcp_port:6379
uptime_in_seconds:3600
uptime_in_days:0
hz:10
configured_hz:10
lru_clock:1234567
executable:/data/DiceDB-server
config_file:/data/DiceDB.conf

# Clients
connected_clients:10
client_longest_output_list:0
client_biggest_input_buf:0
blocked_clients:0

# Memory
used_memory:1024000
used_memory_human:1.00M
used_memory_rss:2048000
used_memory_rss_human:2.00M
used_memory_peak:2048000
used_memory_peak_human:2.00M
used_memory_peak_perc:50.00%
used_memory_overhead:512000
used_memory_startup:512000
used_memory_dataset:512000
used_memory_dataset_perc:50.00%
total_system_memory:4096000
total_system_memory_human:4.00M
used_memory_lua:4096
used_memory_lua_human:4.00K
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
mem_fragmentation_ratio:2.00
mem_allocator:jemalloc-5.1.0

# Persistence
loading:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1620000000
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:0
rdb_current_bgsave_time_sec:-1
aof_enabled:0
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok
aof_current_size:0
aof_base_size:0
aof_pending_rewrite:0
aof_buffer_length:0
aof_rewrite_buffer_length:0
aof_pending_bio_fsync:0
aof_delayed_fsync:0

# Stats
total_connections_received:100
total_commands_processed:1000
instantaneous_ops_per_sec:10
total_net_input_bytes:1024000
total_net_output_bytes:2048000
instantaneous_input_kbps:10.00
instantaneous_output_kbps:20.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
expired_stale_perc:0.00
expired_time_cap_reached_count:0
evicted_keys:0
keyspace_hits:100
keyspace_misses:10
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:0
migrate_cached_sockets:0

# Replication
role:master
connected_slaves:0
master_replid:0000000000000000000000000000000000000000
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:0.00
used_cpu_user:0.00
used_cpu_sys_children:0.00
used_cpu_user_children:0.00

# Commandstats
cmdstat_get:calls=100,usec=1000,usec_per_call=10.00
cmdstat_set:calls=50,usec=500,usec_per_call=10.00

# Cluster
cluster_enabled:0

# Keyspace
db0:keys=10,expires=0,avg_ttl=0
```

### Retrieve Specific Section Information

```bash
127.0.0.1:7379> INFO memory
# Memory
used_memory:1024000
used_memory_human:1.00M
used_memory_rss:2048000
used_memory_rss_human:2.00M
used_memory_peak:2048000
used_memory_peak_human:2.00M
used_memory_peak_perc:50.00%
used_memory_overhead:512000
used_memory_startup:512000
used_memory_dataset:512000
used_memory_dataset_perc:50.00%
total_system_memory:4096000
total_system_memory_human:4.00M
used_memory_lua:4096
used_memory_lua_human:4.00K
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
mem_fragmentation_ratio:2.00
mem_allocator:jemalloc-5.1.0
```

## Behaviour

When the `INFO` command is executed, DiceDB collects and returns the requested information about the server. If no section is specified, it returns all available information. The command does not modify the state of the server or its data; it is purely informational.

## Error Handling

The `INFO` command can raise errors in the following scenarios:

1. `Invalid Section Name`: If an invalid section name is provided, DiceDB will return an error.

   - `Error Message`: `ERR unknown INFO subcommand or wrong number of arguments for 'INFO'`

2. `Syntax Error`: If the command is not used correctly, DiceDB will return a syntax error.

   - `Error Message`: `ERR syntax error`
