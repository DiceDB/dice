Logging Best Practices
===

## Be concise, yet informative
Bad practice:

```go
slog.Info("DiceDB is starting, initialization in progress", slog.String("version", config.DiceDBVersion))
```

Good practice:
```go
slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))
```

## Use structured logging with key-value pairs
Bad practice:
```go
slog.Info("running on port", config.Port)
```

Good practice:
```go
slog.Info("running with", slog.Int("port", config.Port))
```

## Avoid logging redundant information
Bad practice:
```go
slog.Info("running in multi-threaded mode with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))
```

Good practice:
```go
slog.Info("running with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))
```

## Use Boolean values effectively
Bad practice:
```go
slog.Info("enable-watch is set to true", slog.Bool("enable-watch", true))
```

Good practice:
```go
slog.Info("running with", slog.Bool("enable-watch", config.EnableWatch))
```

## Log specific details over general statements
Bad practice:
```go
slog.Info("server is running")
```

Good practice:
```go
slog.Info("running with", slog.Int("port", config.Port), slog.Bool("enable-watch", config.EnableWatch))
```

## Use lowercase for log messages, except for proper nouns
Bad practice:
```go
slog.Info("Starting DiceDB", slog.String("version", config.DiceDBVersion))
```

Good practice:
```go
slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))
```
