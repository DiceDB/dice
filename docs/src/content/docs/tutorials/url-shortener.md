---
title: "Building a URL Shortener"
description: "Create a simple URL Shortener using DiceDB Go SDK."
---

This tutorial guides you through creating a URL shortener using DiceDB, a key-value store, with Go. We’ll set up endpoints to generate short URLs and redirect them to the original URLs.

## Prerequisites

1. Go (version 1.18 or later): [Download Go](https://golang.org/dl/)
2. DiceDB: A DiceDB server running locally. Refer to the [DiceDB Installation Guide](get-started/installation) if you haven't set it up yet.

## Setup

### 1. Install and Run DiceDB

Start a DiceDB server using Docker:

```bash
docker run -d -p 7379:7379 dicedb/dicedb
```

This command pulls the DiceDB Docker image and runs it, exposing it on port `7379`.

### 2. Initialize a New Go Project

Create a new directory for your project and initialize a Go module:

```bash
mkdir url-shortener
cd url-shortener
go mod init url-shortener
```

### 3. Install Required Packages

Install the DiceDB Go SDK and other dependencies:

```bash
go get github.com/dicedb/dicedb-go@v1.0.3
go get github.com/gin-gonic/gin
go get github.com/google/uuid
```

## Understanding DiceDB Commands

We'll use the following DiceDB commands:

### `SET` Command

Stores a key-value pair in DiceDB.

- **Syntax**: `SET key value [expiration]`
  - `key`: Unique identifier (e.g., short URL code)
  - `value`: Data to store (e.g., serialized JSON)
  - `expiration`: Optional; time-to-live in seconds (use `0` for no expiration)

### `GET` Command

Retrieves the value associated with a key.

- **Syntax**: `GET key`
  - `key`: Identifier for the data to retrieve

## Writing the Code

Create a file named `main.go` and add the following code:

- `main.go`:

  ```go
  package main

  import (
      "context"
      "encoding/json"
      "log"
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/google/uuid"
      "github.com/dicedb/dicedb-go" // DiceDB Go SDK
  )

  type URL struct {
      ID       string `json:"id"`
      LongURL  string `json:"long_url"`
      ShortURL string `json:"short_url"`
  }

  var db *dicedb.Client

  // Initialize DiceDB connection
  func init() {
      db = dicedb.NewClient(&dicedb.Options{
          Addr: "localhost:7379",
      })
  }

  // Creates a short URL from a given long URL
  func CreateShortURL(c *gin.Context) {
      var requestBody URL
      if err := c.ShouldBindJSON(&requestBody); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
          return
      }

      // Generate unique short ID and construct the short URL
      shortID := uuid.New().String()[:8]
      requestBody.ID = shortID
      requestBody.ShortURL = "http://localhost:8080/" + shortID

      // Serialize URL struct to JSON and store it in DiceDB
      urlData, err := json.Marshal(requestBody)
      if err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
          return
      }

      if err := db.Set(context.Background(), shortID, urlData, 0).Err(); err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
          return
      }

      c.JSON(http.StatusCreated, gin.H{"short_url": requestBody.ShortURL})
  }

  // Redirects to the original URL based on the short URL ID
  func RedirectURL(c *gin.Context) {
      id := c.Param("id")

      // Retrieve stored URL data from DiceDB
      urlData, err := db.Get(context.Background(), id).Result()
      if err != nil {
          c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
          return
      }

      // Deserialize JSON data back into URL struct
      var url URL
      if err := json.Unmarshal([]byte(urlData), &url); err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode URL data"})
          return
      }

      // Redirect user to the original long URL
      c.Redirect(http.StatusFound, url.LongURL)
  }

  func main() {
      router := gin.Default()

      // Define endpoints for creating short URLs and redirecting
      router.POST("/shorten", CreateShortURL)
      router.GET("/:id", RedirectURL)

      // Start the server on port 8080
      if err := router.Run(":8080"); err != nil {
          log.Fatal("Failed to start server:", err)
      }
  }
  ```

## Explanation

### 1. Initialize the DiceDB Client

We set up the DiceDB client in the `init` function:

```go
db = dicedb.NewClient(&dicedb.Options{
    Addr: "localhost:7379",
})
```

### 2. Create Short URL Endpoint

- **Input Validation**: Ensures the `long_url` field is present.
- **Short ID Generation**: Uses `uuid` to create a unique 8-character ID.
- **Data Serialization**: Converts the `URL` struct to JSON.
- **Data Storage**: Saves the JSON data in DiceDB with the `Set` command.
- **Response**: Returns the generated short URL.

### 3. Redirect to Original URL Endpoint

- **Data Retrieval**: Fetches the URL data from DiceDB using the `Get` command.
- **Data Deserialization**: Converts JSON back to the `URL` struct.
- **Redirection**: Redirects the user to the `LongURL`.

### 4. Start the Server

The `main` function sets up the routes and starts the server on port `8080`.

## Running the Application

### 1. Start the Go Application

```bash
go run main.go
```

This will start the application server on port 8080 by default, you should see output similar to

```bash
[GIN-debug] Listening and serving HTTP on :8080
```

### 2. Ensure DiceDB is Running

Ensure your DiceDB server is up and running on port `7379`.

## Testing the application

### 1. Shorten URL:

**Using `curl`:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{"long_url": "https://example.com"}' http://localhost:8080/shorten
```

**Response:**

```json
{
  "short_url": "http://localhost:8080/<short ID generated by the server>"
}
```

### 2. Redirect to Original URL:

**Using `curl`:**

```bash
curl -L http://localhost:8080/abcd1234
```

**Using a Browser:**
Navigate to:

```
http://localhost:8080/abcd1234
```

You should be redirected to `https://example.com`.
