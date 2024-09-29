# Real-time Leaderboard using DiceDB

This demo showcases a real-time leaderboard application using DiceDB.
Writing realtime applications with DiceDB is as simple as writing a SQL query.

## Overview

The application consists of two main components:

1. A Go backend that interacts with DiceDB and serves WebSocket connections
2. A simple HTML frontend that displays the leaderboard in real-time

The backend randomly generates player scores and updates them in DiceDB using a background thread.
It then uses `QWATCH` to monitor changes and broadcasts updates to connected WebSocket clients.

## Setup Using Docker

```
docker-compose up
```

## Setup on local machine

1. Setup and run DiceDB ([refer](https://github.com/dicedb/dice))
2. Execute the following commands

```
$ cd examples/leaderboard.go
$ go run main.go
```

## See the leaderboard

Open a web browser and navigate to `http://localhost:8000` to view the leaderboard

![DiceDB Leaderboard](https://github.com/user-attachments/assets/327792c7-d788-47d4-a767-ef2c478d75cb)

## Project Structure

- `main.go`: The Go backend application
- `index.html`: The frontend HTML file
- `Dockerfile`: Defines the Docker image for the application
- `docker-compose.yaml`: Compose setup which launches a DiceDB instance and the Go backend

## How it works

1. The Go backend connects to DiceDB and starts updating random player scores every few milliseconds
2. A `QWATCH` query is set up to monitor changes in player scores above a threshold
3. When the query's key-set receives updates, the backend is notified of the updated results
4. The backend then broadcasts the updated leaderboard to all connected WebSocket clients
5. The frontend establishes a WebSocket connection and updates the leaderboard in real-time as it receives data
