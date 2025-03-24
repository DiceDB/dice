# Logging Best Practices

1. Use lowercase for log messages, except for proper nouns

```go
slog.Info("Starting DiceDB", slog.String("version", config.DiceDBVersion))  // not okay
slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))  // okay
```

2. Be concise, yet informative

```go
slog.Info("DiceDB is starting, initialization in progress", slog.String("version", config.DiceDBVersion))  // not okay
slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))  // okay
```

3. Use structured logging with key-value pairs

```go
slog.Info("running on port", config.Port)  // not okay
slog.Info("running with", slog.Int("port", config.Port))  // okay
```

4. Avoid logging redundant information

```go
slog.Info("running in multi-threaded mode with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))  // not okay
slog.Info("running with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))  // okay
```

5. Use Boolean values effectively

```go
slog.Info("enable-watch is set to true", slog.Bool("enable-watch", true))  // not okay
slog.Info("running with", slog.Bool("enable-watch", config.EnableWatch))  // okay
```

6. Log specific details over general statements

```go
slog.Info("server is running")  // not okay
slog.Info("running with", slog.Int("port", config.Port))  // okay
```
