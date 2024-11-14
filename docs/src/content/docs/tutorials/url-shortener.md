---
title: "Building a URL Shortener"
description: "Create a simple URL Shortener using DiceDB Go SDK."
---

This tutorial guides you through creating a URL shortener using DiceDB, a key-value store, with Go. We’ll set up endpoints to generate short URLs and redirect them to the original URLs.

Prerequisites

1. Go installed (at least version 1.18)
2. DiceDB server running locally

## Setup

1. Refer to [DiceDB Installation Guide](get-started/installation) to get your DiceDB server up and running with a simple Docker command.
2. Initialize a New Go Project
3. Install DiceDB Go SDK and other required packges.
    ```bash
    go get github.com/dicedb/dicedb-go
    go get github.com/gin-gonic/gin
    go get github.com/google/uuid
    ```

## DiceDB Commands Used

Here are the main DiceDB commands we’ll use to store and retrieve URLs.

1. `Set` Command: Stores a key-value pair in DiceDB.
Syntax - `Set(key, value, expiration)`
`key`: Unique identifier (e.g., short URL code)
`value`: The data to store (serialized JSON)
`expiration`: Optional; 0 means no expiration

2. `Get` Command: Retrieves the value associated with a key.
Syntax - `Get(key)`
`key`: The identifier for the data to retrieve.

## Code overview

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

## Explanation of Key Parts
1. Database Initialization: The `init()` function sets up a DiceDB client to connect to `localhost:7379`.
2. `CreateShortURL` Endpoint: Handles the `/shorten` route. It generates a unique ID, constructs the short URL, serializes the URL data, and saves it in DiceDB.
3. `RedirectURL` Endpoint: Handles the `/:id` route. It retrieves the original URL by the short ID from DiceDB and redirects the user to it.
4. Starting the Server: The `main` function starts the Gin server on port `8080`.

## Starting the application server

1. Start the application
   ```bash
   go run main.go
   ```
   This will start the application server on port 8080 by default, you should see output similar to
   ```bash
   [GIN-debug] Listening and serving HTTP on :8080
   ```

## Interacting with the application

1. Start DiceDB: Ensure DiceDB is running.
2. Test the API:
    - Shorten URL:
        Send a POST request to `/shorten` with JSON body on Postman:
        ```
        {
            "long_url": "https://example.com"
        }
        ```

        OR

        ```curl
        curl -X POST -H "Content-Type: application/json" -d '{"long_url": "https://example.com"}' http://localhost:8080/shorten
        ```

    - Redirect to Original URL:
        Send a GET request to `/:id` with the short URL ID on Postman

        OR

        ```curl
        curl -L http://localhost:8080/{short_id}
        ```
