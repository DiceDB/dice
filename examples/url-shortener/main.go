// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"encoding/json"
	"os"
	"fmt"
	
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/dicedb/dicedb-go" // DiceDB Go SDK
)

type URL struct {
	LongURL  string `json:"long_url"`
}

var db *dicedb.Client

// Initialize DiceDB connection
func init() {
	dhost := "localhost"
	if val := os.Getenv("DICEDB_HOST"); val != "" {
		dhost = val
	}

	dport := "7379"
	if val := os.Getenv("DICEDB_PORT"); val != "" {
		dport = val
	}

	db = dicedb.NewClient(&dicedb.Options{
		Addr: fmt.Sprintf("%s:%s", dhost, dport),
	})
}

// Creates a short URL from a given long URL
func createShortURL(c *gin.Context) {
	var requestBody URL
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Generate unique short ID and construct the short URL
	shortID := uuid.New().String()[:8]
	shortURL := "http://localhost:8080/" + shortID

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

	c.JSON(http.StatusCreated, gin.H{"short_url": shortURL})
}

// Redirects to the original URL based on the short URL ID
func redirectURL(c *gin.Context) {
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
	router.POST("/shorten", createShortURL)
	router.GET("/:id", redirectURL)

	// Start the server on port 8080
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
