Logging Best Practices
===

1. Be concise, yet informative
   - Bad example: `slog.Info("DiceDB is starting, initialization in progress", slog.String("version", config.DiceDBVersion))`
   - Good example: `slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))`

2. Use structured logging with key-value pairs
   - Bad example: `slog.Info("running on port", config.Port)`
   - Good example: `slog.Info("running with", slog.Int("port", config.Port))`

3. Avoid logging redundant information
   - Bad example: `slog.Info("running in multi-threaded mode with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))`
   - Good example: `slog.Info("running with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))`

4. Use Boolean values effectively
   - Bad example: `slog.Info("enable-watch is set to true", slog.Bool("enable-watch", true))`
   - Good example: `slog.Info("running with", slog.Bool("enable-watch", config.EnableWatch))`

5. Log specific details over general statements
   - Bad example: `slog.Info("server is running")`
   - Good example: `slog.Info("running with", slog.Int("port", config.Port), slog.Bool("enable-watch", config.EnableWatch))`

6. Use lowercase for log messages, except for proper nouns
   - Bad example: `slog.Info("Starting DiceDB", slog.String("version", config.DiceDBVersion))`
   - Good example: `slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))`
